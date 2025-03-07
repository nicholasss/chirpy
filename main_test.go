package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// test helper function
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
		actual := censorString(test.input)
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

// TODO: write a helper function to handle response and conversion to actual responses
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
