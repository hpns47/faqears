package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed all:swagger_assets
var swaggerFS embed.FS

//go:embed openapi.yaml
var openAPISpec []byte

func mountDocs(r chi.Router) {
	sub, err := fs.Sub(swaggerFS, "swagger_assets")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	r.Get("/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(openAPISpec)
	})

	r.Get("/swagger", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", http.StripPrefix("/swagger/", fileServer).ServeHTTP)
}
