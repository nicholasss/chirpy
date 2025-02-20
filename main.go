package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const (
	port = "8080"

	// HTTP Status Codes
	OK = 200

	ServiceUnavailable = 503
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger(r)

		// this closure will be called when a request is processed
		cfg.fileserverHits.Add(1)
		log.Printf("Incremented the fileserver hit counter by 1.")

		// the request is then passed to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerReset(writer http.ResponseWriter, req *http.Request) {
	requestLogger(req)

	writer.WriteHeader(OK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	cfg.fileserverHits.Store(0)
	str := "Reset the fileserver hit counter."
	writer.Write([]byte(str))

	log.Println(str)
}

func (cfg *apiConfig) handlerMetrics(writer http.ResponseWriter, req *http.Request) {
	requestLogger(req)

	writer.WriteHeader(OK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	hits := cfg.fileserverHits.Load()
	str := fmt.Sprintf("Hits: %d", hits)
	writer.Write([]byte(str))

	log.Printf("Served metrics page with %d hits.", hits)
}

// simply logs information about incoming requests
func requestLogger(req *http.Request) {
	agent := req.UserAgent()
	url := req.URL
	host := req.Host
	log.Printf("%s requested %s from %s", agent, url, host)
}

func handlerFS(path string) http.Handler {
	root := http.Dir(".")
	fs := http.FileServer(root)
	handler := http.StripPrefix(path, fs)

	return handler
}

func handlerReady(writer http.ResponseWriter, req *http.Request) {
	requestLogger(req)

	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(OK)
	writer.Write([]byte("OK"))

	log.Printf("Served health page.")
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	// file handler
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handlerFS("/app/")))

	// metrics handler
	mux.Handle("/metrics", http.HandlerFunc(apiCfg.handlerMetrics))

	// reset path
	mux.Handle("/reset", http.HandlerFunc(apiCfg.handlerReset))

	// health/ready handler
	mux.HandleFunc("/healthz", handlerReady)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Listening and Serving on port: '%s'\n", port)
	log.Fatal(server.ListenAndServe()) // server can return error
}
