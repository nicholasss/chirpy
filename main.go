package main

import (
	"fmt"
	"net/http"
)

const (
	port = ":8080"
)

func main() {

	serveMux := http.NewServeMux()

	server := http.Server{
		Addr:    port,
		Handler: serveMux,
	}

	fmt.Printf("Listening and Serving on port: '%s'\n", port)
	server.ListenAndServe()
}
