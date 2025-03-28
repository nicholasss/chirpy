package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMain build up and tear down
// https://markphelps.me/posts/writing-tests-for-your-database-code-in-go/

// func TestMain(m *testing.M) {
// 	// os.Exit skups defer calls
// 	// so we need to call another function
// 	code, err := run(m)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	os.Exit(code)
// }
//
// func run(m *testing.M) (code int, err error) {
// 	// psudo-code
// 	// 1. create test db if it doesnt exist
// 	// 2. run our DDL statements to create required tables
// 	// 3. run our tests
// 	// 4. truncate (???) the test db tables
// }

// Helper functions

// Read and return the Response Status & Response Body
func readResponse(r *httptest.ResponseRecorder, t *testing.T) (string, int) {
	body := r.Result().Body
	defer body.Close()

	bodyData, err := io.ReadAll(body)
	if err != nil {
		t.Errorf("Error was found with respondWithError: %s", err)
	}

	responseBody := string(bodyData)
	responseCode := r.Result().StatusCode

	return responseBody, responseCode
}

// Function Testing

func TestCensorString(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"words", "words"},
		{"no censoring", "no censoring"},
		{"kerfuffle", "****"},
		{"sharbert", "****"},
		{"fornax", "****"},
		{"this is a long sentence", "this is a long sentence"},
	}

	for _, test := range tests {
		actual, err := validateChirp(test.input)
		if err != nil {
			t.Errorf("Body was too long: %s", err)
		}
		if actual != test.expected {
			t.Errorf("Expected '%s', want '%s'", test.expected, actual)
		}
	}
}

func TestNewErrorData(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"Something went wrong", `{"error":"Something went wrong"}`},
		{"Oops!", `{"error":"Oops!"}`},
		{"", `{"error":""}`},
	}

	for _, test := range tests {
		actual := string(newErrorData(test.input))
		if actual != test.expected {
			t.Errorf("Expected '%s', received '%s'", test.expected, actual)
		}
	}
}

func TestRespondWithError(t *testing.T) {
	var tests = []struct {
		inputCode    int
		inputMsg     string
		expectedCode int
		expectedMsg  string
	}{
		{
			http.StatusInternalServerError,
			"Something went wrong",
			http.StatusInternalServerError,
			`{"error":"Something went wrong"}`,
		},
		{
			http.StatusBadGateway,
			"Oopsie whoopsies",
			http.StatusBadGateway,
			`{"error":"Oopsie whoopsies"}`,
		},
		{
			http.StatusGatewayTimeout,
			"Gateway timed out",
			http.StatusGatewayTimeout,
			`{"error":"Gateway timed out"}`,
		},
		{
			http.StatusOK,
			"",
			http.StatusOK,
			`{"error":""}`,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		respondWithError(w, test.inputCode, test.inputMsg)

		actualMsg, actualCode := readResponse(w, t)

		if actualCode != test.expectedCode {
			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, test.inputCode)
		}
		if actualMsg != string(test.expectedMsg) {
			t.Errorf("Expected '%s', recieved '%s'", string(test.expectedMsg), actualMsg)
		}
	}
}

func TestRespondWithJSON(t *testing.T) {
	type Payload struct {
		Str string `json:"str"`
		Num int    `json:"num"`
	}

	var tests = []struct {
		inputCode    int
		intputData   Payload
		expectedCode int
		expectedData string
	}{
		{
			http.StatusOK,
			Payload{
				Str: "test message",
				Num: 1234,
			},
			http.StatusOK,
			`{"str":"test message","num":1234}`,
		},
		{
			http.StatusOK,
			Payload{
				Str: "there will be falafel",
				Num: 420,
			},
			http.StatusOK,
			`{"str":"there will be falafel","num":420}`,
		},
		{
			http.StatusOK,
			Payload{},
			http.StatusOK,
			`{"str":"","num":0}`,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		respondWithJSON(w, test.inputCode, test.intputData)

		actualMsg, actualCode := readResponse(w, t)

		if actualCode != test.expectedCode {
			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, test.inputCode)
		}
		if actualMsg != string(test.expectedData) {
			t.Errorf("Expected '%s', recieved '%s'", string(test.expectedData), actualMsg)
		}
	}
}

func TestHandlerReady(t *testing.T) {
	var tests = []struct {
		expectedMsg  string
		expectedCode int
	}{
		{
			"OK",
			http.StatusOK,
		},
	}

	for _, test := range tests {
		req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
		w := httptest.NewRecorder()
		handlerReady(w, req)

		actualMsg, actualCode := readResponse(w, t)

		if actualCode != test.expectedCode {
			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, actualCode)
		}
		if actualMsg != test.expectedMsg {
			t.Errorf("Expected '%s', received '%s'", test.expectedMsg, actualMsg)
		}
	}
}

// TODO: complete test for the users endpoint
// requires some kind of test database

// func TestHandlerCreateUsers(t *testing.T) {
// 	var tests = []struct {
// 		inputPayload UserCreateRequest
// 		expectedMsg  string
// 		expectedCode int
// 	}{
// 		{
// 			UserCreateRequest{
// 				Email: "tim@apple.com",
// 			},
// 			`{"email":"tim@apple.com"}`,
// 			http.StatusOK,
// 		},
// 	}
//
// 	for _, test := range tests {
// 		dataBuffer := &bytes.Buffer{}
// 		err := json.NewEncoder(dataBuffer).Encode(test.inputPayload)
// 		if err != nil {
// 			t.Fatalf("Unable to encode inputPayload to dataBuffer: %v", err)
// 		}
//
// 		r := httptest.NewRequest(http.MethodPost, "/api/users", dataBuffer)
// 		w := httptest.NewRecorder()
// 		cfg.handlerCreateUsers(w, r)
//
// 		actualMsg, actualCode := readResponse(w, t)
//
// 		if actualCode != test.expectedCode {
// 			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, actualCode)
// 		}
// 		if actualMsg != test.expectedMsg {
// 			t.Errorf("Expected '%s', received '%s'", test.expectedMsg, actualMsg)
// 		}
// 	}
// }

// INFO: old test, no longer needed

// func TestValidateChirp(t *testing.T) {
// 	var tests = []struct {
// 		inputPayload Chirp
// 		expectedMsg  string
// 		expectedCode int
// 	}{
// 		{
// 			Chirp{
// 				Body: "Hozier is so great.",
// 			},
// 			`{"cleaned_body":"Hozier is so great."}`,
// 			http.StatusOK,
// 		},
// 		{
// 			Chirp{
// 				Body: "My fornax is not working.",
// 			},
// 			`{"cleaned_body":"My **** is not working."}`,
// 			http.StatusOK,
// 		},
// 		{
// 			Chirp{
// 				Body: "",
// 			},
// 			`{"cleaned_body":""}`,
// 			http.StatusOK,
// 		},
// 	}
//
// 	for _, test := range tests {
// 		dataBuffer := &bytes.Buffer{}
// 		err := json.NewEncoder(dataBuffer).Encode(test.inputPayload)
// 		if err != nil {
// 			t.Fatalf("Unable to encode inputPayload to dataBuffer: %v", err)
// 		}
//
// 		r := httptest.NewRequest(http.MethodPost, "/api/validate_chirp", dataBuffer)
// 		w := httptest.NewRecorder()
// 		handlerValidate(w, r)
//
// 		actualMsg, actualCode := readResponse(w, t)
//
// 		if actualCode != test.expectedCode {
// 			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, actualCode)
// 		}
// 		if actualMsg != test.expectedMsg {
// 			t.Errorf("Expected '%s', received '%s'", test.expectedMsg, actualMsg)
// 		}
// 	}
// }

func TestHandlerMetrics(t *testing.T) {
	var tests = []struct {
		expectedMsg  string
		expectedCode int
	}{
		{
			`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 0 times!</p>
  </body>
</html>`,
			http.StatusOK,
		},
	}

	cfg := apiConfig{}

	for _, test := range tests {
		r := httptest.NewRequest(http.MethodGet, "/admin/metrics", nil)
		w := httptest.NewRecorder()
		cfg.handlerMetrics(w, r)

		actualMsg, actualCode := readResponse(w, t)

		if actualCode != test.expectedCode {
			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, actualCode)
		}
		if actualMsg != test.expectedMsg {
			t.Errorf("Expected '%s', received '%s'", test.expectedMsg, actualMsg)
		}
	}
}

// func TestHandlerReset(t *testing.T) {
// 	var tests = []struct {
// 		expectedMsg  string
// 		expectedCode int
// 	}{
// 		{
// 			"Reset the fileserver hit counter.",
// 			http.StatusOK,
// 		},
// 	}
//
// 	// by default, if you are testing, you are in dev
// 	cfg := apiConfig{
// 		platform:       "dev",
// 		fileserverHits: atomic.Int32{},
// 		db:             nil,
// 	}
//
// 	for _, test := range tests {
// 		r := httptest.NewRequest(http.MethodPost, "/admin/reset", nil)
// 		w := httptest.NewRecorder()
// 		cfg.handlerReset(w, r)
//
// 		actualMsg, actualCode := readResponse(w, t)
//
// 		if actualCode != test.expectedCode {
// 			t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, actualCode)
// 		}
// 		if actualMsg != test.expectedMsg {
// 			t.Errorf("Expected '%s', received '%s'", test.expectedMsg, actualMsg)
// 		}
// 	}
// }
