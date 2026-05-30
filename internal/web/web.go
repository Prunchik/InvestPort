package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:static
//go:embed all:templates
var embedFS embed.FS

func StaticFileServer() http.Handler {
	sub, err := fs.Sub(embedFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServerFS(sub)
}

func IndexHTML() ([]byte, error) {
	return embedFS.ReadFile("templates/index.html")
}
