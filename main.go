package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

const (
	port = "8080"

	rootPath  = "/"
	appPath   = "/app/"
	readyPath = "/healthz"

	// HTTP Status Codes
	OK = 200

	ServiceUnavailable = 503
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// does a step to count
	cfg.fileserverHits.Add(1)
	log.Printf("incremented the fileserver hit")

	// then returns the handler
	return next
}

func handlerFS() http.Handler {
	root := http.Dir(".")
	fs := http.FileServer(root)

	return http.StripPrefix(appPath, fs)
}

func handlerReady(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(OK)
	writer.Write([]byte("OK"))
}

func main() {

	mux := http.NewServeMux()

	// file handler
	mux.Handle(appPath, handlerFS())

	// health/ready handler
	mux.HandleFunc(readyPath, handlerReady)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Listening and Serving on port: '%s'\n", port)
	log.Fatal(server.ListenAndServe()) // server can return error
}
