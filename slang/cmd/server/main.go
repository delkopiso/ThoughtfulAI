package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(1 * time.Second)
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("processing request with body", buf.String())
		_, _ = writer.Write(buf.Bytes())
	})
	srv := http.Server{Handler: mux, Addr: ":8080"}
	log.Println("starting server at " + srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Errorf("failed to start http listener: %w", err))
	}
}
