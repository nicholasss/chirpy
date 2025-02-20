package main

import (
	"log"
	"net/http"
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

func handlerReady(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(OK)
	writer.Write([]byte("OK"))
}

func main() {

	mux := http.NewServeMux()

	// file handler
	mux.Handle(appPath, http.StripPrefix(appPath, http.FileServer(http.Dir("."))))

	// health/ready handler
	mux.HandleFunc(readyPath, handlerReady)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Listening and Serving on port: '%s'\n", port)
	log.Fatal(server.ListenAndServe()) // server can return error
}
