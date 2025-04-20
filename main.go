package livereload

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

var (
	upgrader = websocket.Upgrader{}
	sessions = map[string]*websocket.Conn{}
)

// watchDir recursively adds all directories in the given root to the watcher
func watchDir(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func handleWebsocketConnection(logger zerolog.Logger) echo.HandlerFunc {

	return func(ctx echo.Context) error {
		ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

		sessionID := ctx.Param("session_id")

		sessions[sessionID] = ws

		if err != nil {
			return err
		}
		defer ws.Close()

		for {
			// Write message
			err := ws.WriteMessage(websocket.TextMessage, []byte("PING"))
			if err != nil {
				break
			}

			// Read message
			_, y, err := ws.ReadMessage()
			logger.Debug().Msgf("[LiveReload] GOT MESSAGE: %v\n", string(y))
			if err != nil {
				// GoodBye
				break
			}
		}

		return nil
	}
}

func handleGetLiveReloadScript() echo.HandlerFunc {

	return func(ctx echo.Context) error {
		html := `
        function guidGenerator() {
            var S4 = function() {
            return (((1+Math.random())*0x10000)|0).toString(16).substring(1);
            };
            return (S4()+S4()+"-"+S4()+"-"+S4()+"-"+S4()+"-"+S4()+S4()+S4());
        }

        let ID = guidGenerator();

        const WS_URL = "/livereload/ws/" + ID;
        let socket;
        let reconnectInterval = 200;
        let maxReconnectInterval = 30000;
        let reconnectAttempts = 0;
        let maxReconnectAttempts = 10;

        function connectWebSocket() {
            console.log("Connecting to WebSocket...");
            socket = new WebSocket(WS_URL);

            socket.onopen = function () {
                console.log("[open] Connection established");
                socket.send("WAKEUP FROM " + guidGenerator());

                // Reset reconnection attempts on successful connection
                reconnectAttempts = 0;
                reconnectInterval = 1000;
            };

            socket.onmessage = function (event) {
                console.log("Received message: ", event.data);
                if (event.data === "RELOAD") {
                    socket.send("RELOADING FROM " + guidGenerator());
                    location.reload();
                }
            };

            socket.onclose = function (event) {
                if (event.wasClean) {
                    console.log("WebSocket Connection closed.");
                } else {
                    console.warn("WebSocket Connection lost, attempting to reconnect...");
                    reconnectWebSocket();
                }
            };

            socket.onerror = function (error) {
                console.error("WebSocket Error:", error);
            };
        }

        function reconnectWebSocket() {
            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                let timeout = Math.min(reconnectInterval * 2, maxReconnectInterval); // Exponential backoff

                console.log("Reconnecting in " + (timeout / 1000) + " seconds (attempt " + reconnectAttempts + "/" + maxReconnectAttempts + ")...");
                
                setTimeout(() => {
                    connectWebSocket();
                }, timeout);

                reconnectInterval = timeout;
            } else {
                console.error("WebSocket Error Max reconnect attempts reached. No more reconnecting.");
            }
        }

        connectWebSocket();
        `
		return ctx.HTML(200, html)
	}
}

func LiveReload(echoServer *echo.Echo, logger zerolog.Logger, dirs ...string) echo.MiddlewareFunc {

	echoServer.GET("/livereload/ws/:session_id", handleWebsocketConnection(logger))

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		panic("livereload: " + err.Error())
	}

	for _, dir := range dirs {
		watchDir(watcher, dir)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.Debug().Msgf("[LiveReload] File %s changed\n", event.Name)

					time.Sleep(100 * time.Millisecond)
					var sessionToRemove []string
					for sessionID, websocketConn := range sessions {
						message := []byte("RELOAD")
						err := websocketConn.WriteMessage(1, message)

						if err != nil {
							sessionToRemove = append(sessionToRemove, sessionID)
							logger.Debug().Msgf("[LiveReload] Removing dangling websocket connection: %s\n", sessionID)
						}
					}

					for _, x := range sessionToRemove {
						delete(sessions, x)
					}

				} else if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						// Watch newly created directories
						if err := watcher.Add(event.Name); err == nil {
							logger.Debug().Msgf("New directory added to watcher: %s\n", event.Name)
						}
					}
				}

			}
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			path := c.Request().URL.Path
			if path == "/livereload.js" {
				return handleGetLiveReloadScript()(c)
			}
			return next(c)
		}
	}
}
