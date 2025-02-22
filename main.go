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

func (cfg *apiConfig) mwLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL.Path)
		next.ServeHTTP(writer, req)
	})
}

func (cfg *apiConfig) mwMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		// this closure will be called when a request is processed
		cfg.fileserverHits.Add(1)
		log.Printf("Incremented the fileserver hit counter by 1.")

		// the request is then passed to the next handler in the chain
		next.ServeHTTP(writer, req)
	})
}

func (cfg *apiConfig) handlerReset(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(OK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	cfg.fileserverHits.Store(0)
	str := "Reset the fileserver hit counter."
	writer.Write([]byte(str))

	log.Println(str)
}

func (cfg *apiConfig) handlerMetrics(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(OK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	hits := cfg.fileserverHits.Load()
	str := fmt.Sprintf("Hits: %d", hits)
	writer.Write([]byte(str))

	log.Printf("Served metrics page with %d hits.", hits)
}

func handlerFS(path string) http.Handler {
	root := http.Dir(".")
	fs := http.FileServer(root)
	handler := http.StripPrefix(path, fs)

	return handler
}

func handlerReady(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(OK)
	writer.Write([]byte("OK"))

	log.Printf("Served health page.")
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	// file handler
	mux.Handle("/app/", apiCfg.mwLog(apiCfg.mwMetricsInc(handlerFS("/app/"))))

	// metrics handler
	mux.Handle("GET /metrics", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerMetrics)))

	// reset path
	mux.Handle("POST /reset", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerReset)))

	// health/ready handler
	mux.Handle("GET /healthz", apiCfg.mwLog(http.HandlerFunc(handlerReady)))

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Listening and Serving on port: '%s'\n", port)
	log.Fatal(server.ListenAndServe()) // server can return error
}
