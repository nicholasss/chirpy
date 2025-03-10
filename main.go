package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/google/uuid"
	_ "github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nicholasss/chirpy/internal/database"
)

// =========
// CONSTANTS
// =========

const (
	port          = "8080"
	maxChirpRunes = 140
)

// ================
// GLOBAL VARIABLES
// ================

// admin metrics page
// %d needs to be replaced with the number of hits
var adminMetricsPage = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

// words that need to be censored in the chirps
var censoredWords = []string{"kerfuffle", "sharbert", "fornax"}

// ============
// GLOBAL TYPES
// ============

type apiConfig struct {
	platform       string
	fileserverHits atomic.Int32
	db             *database.Queries
}

// API types

type UserCreateRequest struct {
	Email string `json:"email"`
}
type Chirp struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}
type CleanedChirp struct {
	CleanedBody string    `json:"cleaned_body"`
	UserID      uuid.UUID `json:"user_id"`
}

// Internal types

type ErrorResponse struct {
	Error string `json:"error"`
}
type ValidResponse struct {
	Valid bool `json:"valid"`
}

// =================
// UTILITY FUNCTIONS
// =================

// censors the following words: kerfuffle, sharbert, fornax
// replaces them with **** (four asterisks)
func validateChirp(text string) (string, error) {
	chirpLen := utf8.RuneCountInString(text)
	if chirpLen >= maxChirpRunes {
		fmt.Printf("Chirp too long: %d, %d chars too many.\n", chirpLen, maxChirpRunes-chirpLen)
		return "", fmt.Errorf("chirp is too long. %d chars too many\n", maxChirpRunes-chirpLen)
	}

	cleanedWords := make([]string, 0)
	words := strings.Split(text, " ")

	for _, word := range words {
		testWord := strings.ToLower(word)
		if slices.Contains(censoredWords, testWord) {
			cleanedWords = append(cleanedWords, "****")
			continue
		}
		cleanedWords = append(cleanedWords, word)
	}

	censoredString := strings.Join(cleanedWords, " ")
	return censoredString, nil
}

func newErrorData(cause string) []byte {
	errorRecord := ErrorResponse{Error: cause}
	errorData, err := json.Marshal(errorRecord)
	if err != nil {
		log.Fatalf("Unable to encode error response: %s", err)
	}

	return errorData
}

// responds to request with an error specified
func respondWithError(w http.ResponseWriter, code int, msg string) {
	errorData := newErrorData(msg)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(errorData)
}

// response to request with a json payload, specified
func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Unable to encode valid response: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(payloadData)
}

// ====================
// MIDDLEWARE FUNCTIONS
// ====================

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

// =============
// HANDLER TYPES
// =============

func handlerFS(path string) http.Handler {
	root := http.Dir(".")
	fs := http.FileServer(root)
	handler := http.StripPrefix(path, fs)

	return handler
}

// =================
// HANDLER FUNCTIONS
// =================

// create chrips with a specified user uuid
func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	var createChirpRequest Chirp
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&createChirpRequest)
	if err != nil {
		log.Printf("Error decoding create chirp request: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
	}

	// 1. requires body and user_id fields
	// TODO: Verify that user ID links to a valid user
	if err = uuid.Validate(createChirpRequest.UserID.String()); err != nil {
		log.Print("Create Chirp request has either body or user_id missing.")
		respondWithError(w, http.StatusBadRequest, "Unable to validate user uuid.")
	}

	// 2. validate the body and censor strings
	validBody, err := validateChirp(createChirpRequest.Body)
	if err != nil {
		log.Printf("Chirp is too long. %s\n", err)
		respondWithError(w, http.StatusBadRequest, "Chirp is too long.")
	}
	createChirpRequest.Body = validBody

	// 3. insert into database
	chirpRecord, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   createChirpRequest.Body,
		UserID: createChirpRequest.UserID,
	})
	if err != nil {
		log.Printf("Chirp table error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
	}

	// 4. respond with a 201 (status created) and the full record
	log.Print("Processed create chirp successfuly.")
	respondWithJSON(w, http.StatusCreated, chirpRecord)

}

// creates users with a specified email
func (cfg *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {
	var createUserRecord UserCreateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&createUserRecord)
	if err != nil {
		log.Printf("Error decoding create user request: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	userRecord, err := cfg.db.CreateUser(r.Context(), createUserRecord.Email)
	if err != nil {
		// TODO: could do some db err decoding here
		log.Printf("Error creating new user record: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	log.Printf("New user created with '%s'.", userRecord.Email)
	respondWithJSON(w, http.StatusCreated, userRecord)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	hits := cfg.fileserverHits.Load()
	str := fmt.Sprintf(adminMetricsPage, hits)
	w.Write([]byte(str))

	log.Printf("Served metrics page with %d hits.", hits)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	var reset string
	var users string

	if cfg.platform == "production" {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Forbidden.\n"))
		return

	} else if cfg.platform == "dev" {
		// resets user database
		cfg.db.ResetUsers(r.Context())
		users = "Reset Users table.\n"
		log.Print(users)

		// reset hit counter for /api/*
		cfg.fileserverHits.Store(0)
		reset = "Reset hit counter.\n"
		log.Print(reset)

	} else {
		log.Fatal("Platform is not set in ./.env")
		return
	}

	buffer := reset + users
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(buffer))
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	log.Printf("Served health page.")
}

// ====
// MAIN
// ====

func main() {
	// setting up connection to the database
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Unable to load '.env'.")
	}

	platform := os.Getenv("PLATFORM")

	dbURL := os.Getenv("GOOSE_DBSTRING")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to open connection to database.")
	}

	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		platform:       platform,
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	mux := http.NewServeMux()

	// generic endpoints
	mux.Handle("/app/", apiCfg.mwLog(apiCfg.mwMetricsInc(handlerFS("/app/"))))

	// API endpoints
	mux.Handle("GET /api/healthz", apiCfg.mwLog(http.HandlerFunc(handlerReady)))
	mux.Handle("POST /api/users", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerCreateUsers)))
	mux.Handle("POST /api/chirps", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerCreateChirps)))

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
