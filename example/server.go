package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/sevaho/livereload"
)

func main() {
	e := echo.New()
	e.Use(livereload.LiveReload(e, log.Logger, "."))

	templates := NewHtmlEngine("templates")

	e.GET("/", func(c echo.Context) error {

		return c.HTMLBlob(http.StatusOK, templates.RenderHTML("index", nil))

	}).Name = "foobar"

	e.Logger.Fatal(e.Start(":8000"))
}
