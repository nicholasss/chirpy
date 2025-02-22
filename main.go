package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"unicode/utf8"
)

const (
	port          = "8080"
	maxChirpRunes = 140
)

var adminMetricsPage = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) mwLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) mwMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// this closure will be called when a request is processed
		cfg.fileserverHits.Add(1)
		log.Printf("Incremented the fileserver hit counter by 1.")

		// the request is then passed to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	cfg.fileserverHits.Store(0)
	str := "Reset the fileserver hit counter."
	w.Write([]byte(str))

	log.Println(str)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	hits := cfg.fileserverHits.Load()
	str := fmt.Sprintf(adminMetricsPage, hits)
	w.Write([]byte(str))

	log.Printf("Served metrics page with %d hits.", hits)
}

func handlerFS(path string) http.Handler {
	root := http.Dir(".")
	fs := http.FileServer(root)
	handler := http.StripPrefix(path, fs)

	return handler
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	log.Printf("Served health page.")
}

// Accepts POST and expects a json object of a particulate shape
func handlerValidate(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body string `json:"body"`
	}
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	type ValidResponse struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpRecord := Chirp{}
	err := decoder.Decode(&chirpRecord)
	if err != nil {
		log.Printf("Error decoding chirp record: %s", err)

		errorRecord := ErrorResponse{Error: "Something went wrong"}
		errorData, err := json.Marshal(errorRecord)
		if err != nil {
			log.Fatalf("Unable to encode error response: %s", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errorData)
		return
	}

	// validate the chirp
	chirpLen := utf8.RuneCountInString(chirpRecord.Body)
	if maxChirpRunes > chirpLen {
		log.Println("Valid Chirp processed.")

		validRecord := ValidResponse{Valid: true}
		validData, err := json.Marshal(validRecord)
		if err != nil {
			log.Fatalf("Unable to encode valid response: %s", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(validData)
		return
	}

	log.Println("Invalid Chirp processed.")
	errorRecord := ErrorResponse{Error: "Chirp is too long"}
	errorData, err := json.Marshal(errorRecord)
	if err != nil {
		log.Fatalf("Unable to encode error response: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(errorData)
	return
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	// generic endpoints
	mux.Handle("/app/", apiCfg.mwLog(apiCfg.mwMetricsInc(handlerFS("/app/"))))

	// API endpoints
	mux.Handle("GET /api/healthz", apiCfg.mwLog(http.HandlerFunc(handlerReady)))
	mux.Handle("POST /api/validate_chirp", apiCfg.mwLog(http.HandlerFunc(handlerValidate)))

	// Admin endpoints
	mux.Handle("GET /admin/metrics", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerMetrics)))
	mux.Handle("POST /admin/reset", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerReset)))

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Listening and Serving on port: '%s'\n", port)
	log.Fatal(server.ListenAndServe()) // server can return error
}
