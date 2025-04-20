package main

import (
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/valyala/bytebufferpool"

	"github.com/gofiber/template/html/v2"
)

var engine *html.Engine

type HtmlEngine struct {
	htmlEngine *html.Engine
}

func NewHtmlEngine(directory string) *HtmlEngine {
	engine = html.NewFileSystem(http.Dir(directory), ".html")

	engine.Reload(true)

	return &HtmlEngine{htmlEngine: engine}
}

func (e *HtmlEngine) render(w io.Writer, name string, data any, layout string) error {
	return engine.Render(w, name, data, layout)
}

func (e *HtmlEngine) RenderHTML(name string, data map[string]any, layout ...string) []byte {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	if len(layout) > 0 {
		if err := e.render(buf, name, data, layout[0]); err != nil {
			log.Logger.Error().Err(err).Msg("Something went wrong while rendering HTML layouts.")
			panic(err)
		}
	} else {
		if err := e.render(buf, name, data, ""); err != nil {

			log.Logger.Error().Err(err).Msg("Something went wrong while rendering HTML layouts.")
			panic(err)
		}
	}

	return buf.Bytes()
}
