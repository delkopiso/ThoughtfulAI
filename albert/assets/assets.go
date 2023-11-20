package assets

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed public
var public embed.FS

func Mount(router chi.Router) {
	router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.FS(public))))
}
