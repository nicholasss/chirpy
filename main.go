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
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nicholasss/chirpy/internal/auth"
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
	jwtSecret      string
}

// API types

type UserLoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}
type UserLoginRequest struct {
	RawPassword string `json:"password"`
	Email       string `json:"email"`
}
type UserCreateRequest struct {
	RawPassword string `json:"password"`
	Email       string `json:"email"`
}
type UserUpdateRequest struct {
	RawPassword string `json:"password"`
	Email       string `json:"email"`
}
type UserUpgradeRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

// non-user

type AccessTokenResponse struct {
	AccessToken string `json:"token"`
}
type Chirp struct {
	Body string `json:"body"`
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
		return "", fmt.Errorf("chirp is too long. %d chars too many", maxChirpRunes-chirpLen)
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
		return
	}

	// get JWT from headers
	requestToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting request token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// validate token
	userIDFromToken, err := auth.ValidateJWT(requestToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error validating request token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// get uuid from database based on token
	userRecord, err := cfg.db.GetUserByIDSafe(r.Context(), userIDFromToken)
	if err != nil {
		log.Printf("Error validating UUID from token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// checking token validation and registered user
	if userIDFromToken != userRecord.ID {
		log.Printf("Invalid JWT was presented. Expected %s, Got from db %s", userIDFromToken, userRecord.ID)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 1. requires body and user_id fields
	if err = uuid.Validate(userRecord.ID.String()); err != nil {
		log.Print("Create Chirp request has user_id missing.")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. validate the body and censor strings
	validBody, err := validateChirp(createChirpRequest.Body)
	if err != nil {
		log.Printf("Chirp is too long. %s\n", err)
		respondWithError(w, http.StatusBadRequest, "Chirp is too long.")
		return
	}
	createChirpRequest.Body = validBody

	// 3. insert into database
	chirpRecord, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   createChirpRequest.Body,
		UserID: userRecord.ID,
	})
	if err != nil {
		log.Printf("Chirp table error: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// 4. respond with a 201 (status created) and the full record
	log.Print("Processed create chirp successfuly.")
	respondWithJSON(w, http.StatusCreated, chirpRecord)
}

// optional query param of 'author_id' will return only that authors chirps
// otherwise, return all as normal
func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	// check if query param exists
	authorID := r.URL.Query().Get("author_id")
	authorToSearch, err := uuid.Parse(authorID)
	if err != nil {
		log.Printf("Error parsing author_id from passed param: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// if there is no query param:
	if authorID == "" {
		chirpRecords, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("Error performing all chirps request: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
			return
		}

		log.Print("Providing response with all chirps.")
		respondWithJSON(w, http.StatusOK, chirpRecords)
	} else {
		// looking for specific author
		chirpRecords, err := cfg.db.GetAllChirpsByAuthorID(r.Context(), authorToSearch)
		if err != nil {
			log.Printf("Error performing all chirps by id request: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
			return
		}

		log.Printf("Providing response with all chirps by author: %s", authorToSearch)
		respondWithJSON(w, http.StatusOK, chirpRecords)
	}
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.Printf("Error parsing uuid in GET URL. Got='%s', %s", r.PathValue("id"), err)
		respondWithError(w, http.StatusInternalServerError, "Invalid ID")
		return
	}

	chirpRecord, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Chirp not found by ID: %s", err)
		respondWithError(w, http.StatusNotFound, "Chirp not found.")
		return // needs to return after error?
	}

	log.Printf("Providing response with chirp id: %s", chirpID.String())
	respondWithJSON(w, http.StatusOK, chirpRecord)
}

// delete a chirp by id with authentication and authorization
func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting request token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	tokenUUID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error validating UUID from token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	// requestor has a valid JWT

	chirpID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.Printf("Unable to parse UUID in DELETE URL. Got='%v', %s", r.PathValue("id"), err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpRecord, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Chirp not found by ID: %s", err)
		respondWithError(w, http.StatusNotFound, "Chirp not found.")
		return // needs to return after error?
	}

	if tokenUUID != chirpRecord.UserID {
		log.Printf("Unable to delete chirp with unauthorized user '%s'", tokenUUID)
		respondWithError(w, http.StatusForbidden, "Unauthorized")
		return
	}

	// user has been authenticated and is authorized to delete chirp
	err = cfg.db.DeleteChirpByID(r.Context(), chirpRecord.ID)
	if err != nil {
		log.Printf("Unable to delete chirp by id '%s'. Error: %s", chirpRecord.ID, err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// chirp was deleted and user was authorized
	log.Printf("Chirp ID '%s' was successfully deleted by '%s'", chirpRecord.ID, tokenUUID)
	w.WriteHeader(http.StatusNoContent)
}

// accepts refresh token in header as authentication
// it should respond with a new jwt access token if authorized
func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Could not find refresh token in auth header: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// check refreshToken in the db
	refreshTokenRecord, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Could not find refresh token in database: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	refreshTokenUserID := refreshTokenRecord.UserID
	refreshTokenExpiry := refreshTokenRecord.ExpiresAt
	refreshTokenRevocation := refreshTokenRecord.RevokedAt

	// check revocation
	if refreshTokenRevocation.Valid {
		// has been revoked
		revocationTime := refreshTokenRevocation.Time
		if time.Now().UTC().After(revocationTime) {
			log.Printf("Refresh token sent to POST /api/refresh is revoked.")
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// has been marked to be revoked, but in the future
		// these tokens should not be accepted:
		// tokens marked to be revoked in the future is not possible
		// this may preset a logical bug
		log.Print("!!! potential bug, check POST /api/refresh handler")
		log.Print("Refresh token will be revoked in the future.")
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// check expiry
	if time.Now().UTC().After(refreshTokenExpiry) {
		log.Printf("Refresh token sent to POST /api/refresh is expired")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// only valid refresh tokens remain
	// create new access token
	accessTokenExpiry := time.Duration(time.Hour * 1)
	newAccessToken, err := auth.MakeJWT(refreshTokenUserID, cfg.jwtSecret, accessTokenExpiry)
	if err != nil {
		log.Printf("Unable to make new access token (jwt): %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	newAccessTokenResponse := AccessTokenResponse{newAccessToken}
	respondWithJSON(w, http.StatusOK, newAccessTokenResponse)
}

// revoke refresh token that matches what was passed in
func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Could not find refresh token in auth header: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// check refresh token table
	err = cfg.db.RevokeRefreshTokenWithToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Database does not contain submitted refresh token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// token was revoked
	// respond with 204, no content (body)
	w.WriteHeader(http.StatusNoContent)
}

// logs in with a specified email and password
// should return a refresh token, as well as a jwt token
func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// Decoding request json
	var loginUserRecord UserLoginRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginUserRecord)
	if err != nil {
		log.Printf("Error decoding create user request: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// checking password hashes
	unsafeUserRecord, err := cfg.db.GetUserByEmailRetHashedPassword(r.Context(), loginUserRecord.Email)
	if err != nil {
		log.Printf("Error getting user record by email: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	err = auth.CheckPasswordHash(loginUserRecord.RawPassword, unsafeUserRecord.HashedPassword)
	if err != nil {
		log.Printf("User login with wrong password attempted for '%s'", loginUserRecord.Email)
		respondWithError(w, http.StatusUnauthorized, "Wrong email or password.")
		return
	}
	// set raw password to zeroval, now that we have verified it
	loginUserRecord.RawPassword = ""

	// retrieve userRecord without password
	safeUserRecord, err := cfg.db.GetUserByEmailSafe(r.Context(), loginUserRecord.Email)
	if err != nil {
		log.Printf("Error getting user record by email: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// generate jwt token for user with 1 hour accessTokenExpiry
	durationHour := time.Duration(time.Hour * 1)
	accessToken, err := auth.MakeJWT(safeUserRecord.ID, cfg.jwtSecret, durationHour)
	if err != nil {
		log.Printf("Error making JWT: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// generate refresh token for user
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error making refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// add refresh token to database which expires in 60 days
	sixtyDayExpiry := time.Duration(time.Hour * 24 * 60)
	refreshTokenExpiry := time.Now().UTC().Add(sixtyDayExpiry)
	cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		ID:        refreshToken,
		UserID:    safeUserRecord.ID,
		ExpiresAt: refreshTokenExpiry,
	})

	// send response and log it
	loginResponseRecord := UserLoginResponse{
		ID:           safeUserRecord.ID,
		CreatedAt:    safeUserRecord.CreatedAt,
		UpdatedAt:    safeUserRecord.UpdatedAt,
		Email:        safeUserRecord.Email,
		IsChirpyRed:  safeUserRecord.IsChirpyRed,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	log.Printf("User '%s' logged in successfuly.", safeUserRecord.Email)
	respondWithJSON(w, http.StatusOK, loginResponseRecord)
}

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	var userUpgradeRequest UserUpgradeRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userUpgradeRequest)
	if err != nil {
		log.Printf("Error decoding create user request: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// check for api key
	keyString, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("Error finding API key in webhook request: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// check api key
	envAPIKey := os.Getenv("POLKA_KEY")
	if keyString != envAPIKey {
		log.Printf("webhook request used invalid API Key: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// api key is valid

	// check for user.upgraded event
	eventType := userUpgradeRequest.Event
	if eventType != "user.upgraded" {
		log.Print("recieved event of non-user.upgraded type")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// event is user.upgraded

	userID := userUpgradeRequest.Data.UserID
	err = cfg.db.UpgradeUserByID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Printf("User id: '%s' was upgraded", userID)
	w.WriteHeader(http.StatusNoContent)
}

// creates users with a specified email
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest UserCreateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&createUserRequest)
	if err != nil {
		log.Printf("Error decoding create user request: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.")
		return
	}

	// ensure there is a password
	if createUserRequest.RawPassword == "" {
		log.Print("Create user request did not have provided password.")
		respondWithError(w, http.StatusBadRequest, "Please try to create your account again.")
		return
	}

	hashedPassword, err := auth.HashPassword(createUserRequest.RawPassword)
	if err != nil {
		log.Printf("Error hashing provided password: %s", err)
		respondWithError(w, http.StatusBadRequest, "Please try to create your account again.")
		return
	}
	createUserRequest.RawPassword = ""

	userRecord, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          createUserRequest.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		// TODO: could do some db err decoding here
		log.Printf("Error creating new user record: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// response to creating account
	// need to POST /api/login for access token and refresh token
	safeUserRecord := UserLoginResponse{
		ID:           userRecord.ID,
		CreatedAt:    userRecord.CreatedAt,
		UpdatedAt:    userRecord.UpdatedAt,
		Email:        userRecord.Email,
		IsChirpyRed:  userRecord.IsChirpyRed,
		AccessToken:  "",
		RefreshToken: "",
	}

	log.Printf("New user created with '%s'.", userRecord.Email)
	respondWithJSON(w, http.StatusCreated, safeUserRecord)
}

// updates users email or password using credential
func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// check access token
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("No authentication token was presented: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	tokenUUID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Unable to validate presented token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// decoding body json to struct
	userUpdateRequest := UserUpdateRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&userUpdateRequest)
	if err != nil {
		log.Printf("Unable to decode body to json: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	safeUserRecord, err := cfg.db.GetUserByIDSafe(r.Context(), tokenUUID)
	if err != nil {
		// token is valid with the user ID not in database is an issue
		log.Printf("!!! potential bug, check PUT /api/users")
		log.Printf("Unable to find email in user database: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// hash password for storage
	newHashedPassword, err := auth.HashPassword(userUpdateRequest.RawPassword)
	if err != nil {
		log.Printf("Unable to hash presented new password")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	newEmail := userUpdateRequest.Email

	updatedSafeUserRecord, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             safeUserRecord.ID,
		Email:          newEmail,
		HashedPassword: newHashedPassword,
	})
	if err != nil {
		log.Printf("Unable to complete user update for id '%s'", safeUserRecord.ID.String())
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// updated successfuly
	log.Printf("Updated user id '%s' successfuly with new email and password", safeUserRecord.ID.String())
	respondWithJSON(w, http.StatusOK, updatedSafeUserRecord)
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

	switch cfg.platform {
	case "production":
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Forbidden.\n"))
		return

	case "development":
		// resets user database
		cfg.db.ResetUsers(r.Context())
		users = "Reset Users table.\n"
		log.Print(users)

		// reset hit counter for /api/*
		cfg.fileserverHits.Store(0)
		reset = "Reset hit counter.\n"
		log.Print(reset)

	default:
		log.Printf("Unknown platform. Please use either 'production' or 'development'.")
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
	// loading from .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Unable to load '.env'.")
	}

	// platform info
	// 'development' or 'production'
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("Unable to load platform key. Please check the README.md.")
	}

	// DB connection info
	dbURL := os.Getenv("GOOSE_DBSTRING")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to open connection to database.")
	}
	dbQueries := database.New(db)

	// JWT secret string
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("Unable to load JWT token. Proceding would be insecure.")
	}

	apiCfg := &apiConfig{
		platform:       platform,
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		jwtSecret:      JWTSecret,
	}

	mux := http.NewServeMux()

	// generic endpoints
	mux.Handle("/app/", apiCfg.mwLog(apiCfg.mwMetricsInc(handlerFS("/app/"))))
	mux.Handle("GET /api/healthz", apiCfg.mwLog(http.HandlerFunc(handlerReady)))

	// users endpoints
	mux.Handle("POST /api/users", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerCreateUser)))
	mux.Handle("PUT /api/users", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerUpdateUser)))
	mux.Handle("POST /api/login", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerLoginUser)))
	mux.Handle("POST /api/polka/webhooks", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerUpgradeUser)))

	// refresh token specific
	mux.Handle("POST /api/refresh", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerRefresh)))
	mux.Handle("POST /api/revoke", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerRevoke)))

	// chirp endpoints
	mux.Handle("POST /api/chirps", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerCreateChirps)))
	mux.Handle("GET /api/chirps", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerGetAllChirps)))
	mux.Handle("GET /api/chirps/{id}", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerGetChirpByID)))
	mux.Handle("DELETE /api/chirps/{id}", apiCfg.mwLog(http.HandlerFunc(apiCfg.handlerDeleteChirpByID)))

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
