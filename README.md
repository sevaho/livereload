# Livereload

> Livereload your golang web app to increase the development experience.

## Install

```
go get github.com/sevaho/livereload
```


## How to use

```golang
import github.com/labstack/echo/v4

server := echo.New()

server.use(livereload.Livereload(server, zerolog.Logger, ...directories_to_watch))
```

And add the following to every HTML file or to a `layout` html file:

```html
<script src="/livereload.js"></script>
```

