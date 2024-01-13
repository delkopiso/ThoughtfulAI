package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

const defaultPort = 8080
const folderNameParam = "folder"

//go:embed index.gohtml
var indexHtml embed.FS

type ViewContainer struct {
	FolderName string
	Entries    []string
}

func indexHandler() http.HandlerFunc {
	indexTemplate := template.Must(template.ParseFS(indexHtml, "index.gohtml"))

	return func(writer http.ResponseWriter, request *http.Request) {
		var viewContainer ViewContainer
		viewContainer.FolderName = request.FormValue(folderNameParam)

		err := filepath.WalkDir(viewContainer.FolderName, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("failed to process %q: %w", path, err)
			}
			if path != viewContainer.FolderName {
				filename := filepath.Base(path)
				viewContainer.Entries = append(viewContainer.Entries, filename)
			}
			return nil
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)
		_ = indexTemplate.ExecuteTemplate(writer, "index.gohtml", viewContainer)
	}
}

func main() {
	addr := fmt.Sprintf("0.0.0.0:%d", defaultPort)
	router := http.NewServeMux()
	router.Handle("/", indexHandler())
	if err := http.ListenAndServe(addr, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Errorf("failed to start server: %w", err))
	}
}
