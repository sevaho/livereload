# Livereload

> Livereload your golang echo labstack web app to increase the development experience. This library only works with [echo](https://echo.labstack.com/) framework.


## Install

```
go get github.com/sevaho/livereload
```


## How to use

```golang
import "github.com/labstack/echo/v4"

server := echo.New()

server.use(livereload.Livereload(server, zerolog.Logger, ...directories_to_watch))
```

And add the following to every HTML file or to a `layout` html file:

```html
<script src="/livereload.js"></script>
```

## Check out the example!

![demo.mp4](.github/blobs/demo.mp4)
