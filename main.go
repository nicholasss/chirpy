package main

import (
	"log"
	"net/http"
)

const (
	port = ":8080"
	root = "/"
)

func main() {

	serveMux := http.NewServeMux()
	serveMux.Handle(root, http.FileServer(http.Dir(".")))

	server := http.Server{
		Addr:    port,
		Handler: serveMux,
	}

	log.Printf("Listening and Serving on port: '%s'\n", port)
	server.ListenAndServe()
}
