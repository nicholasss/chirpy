package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		expected []byte
	}{
		{"Something went wrong", []byte(`{"error":"Something went wrong"}`)},
		{"Oops!", []byte(`{"error":"Oops!"}`)},
	}

	for _, test := range tests {
		actual := newErrorData(test.input)
		if string(actual) != string(test.expected) {
			t.Errorf("Expected '%s', received '%s'", string(test.expected), string(actual))
		}
	}
}

// TODO: write a helper function to handle response and conversion to actual responses
func TestRespondWithError(t *testing.T) {
	var test = struct {
		inputCode    int
		inputMsg     string
		expectedCode int
		expectedMsg  []byte
	}{
		http.StatusInternalServerError,
		"Something went wrong",
		http.StatusInternalServerError,
		[]byte(`{"error":"Something went wrong"}`),
	}

	w := httptest.NewRecorder()
	respondWithError(w, test.inputCode, test.inputMsg)

	res := w.Result()
	defer res.Body.Close()
	bodyData, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error was found with respondWithError: %s", err)
	}

	actualCode := res.StatusCode
	actualMsg := string(bodyData)

	if actualCode != test.expectedCode {
		t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, test.inputCode)
	}
	if actualMsg != string(test.expectedMsg) {
		t.Errorf("Expected '%s', recieved '%s'", string(test.expectedMsg), actualMsg)
	}
}

func TestRespondWithJSON(t *testing.T) {
	type PayloadType struct {
		Str string `json:"str"`
		Num int    `json:"num"`
	}
	payload := PayloadType{
		"test message",
		1234,
	}

	var test = struct {
		inputCode    int
		intputData   PayloadType
		expectedCode int
		expectedData []byte
	}{
		http.StatusOK,
		payload,
		http.StatusOK,
		[]byte(`{"str":"test message","num":1234}`),
	}

	w := httptest.NewRecorder()
	respondWithJSON(w, test.inputCode, test.intputData)

	res := w.Result()
	defer res.Body.Close()
	bodyData, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error was found with respondWithError: %s", err)
	}

	actualCode := res.StatusCode
	actualMsg := string(bodyData)

	if actualCode != test.expectedCode {
		t.Errorf("Expected '%d', recieved '%d'", test.expectedCode, test.inputCode)
	}
	if actualMsg != string(test.expectedData) {
		t.Errorf("Expected '%s', recieved '%s'", string(test.expectedData), actualMsg)
	}
}

func TestHandlerReady(t *testing.T) {
	expected := "OK"

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	handlerReady(w, req)

	res := w.Result()
	defer res.Body.Close()
	bodyData, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error was found with handlerReady: %s", err)
	}

	actual := string(bodyData)

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected '%d', recieved '%d'", http.StatusOK, res.StatusCode)
	}
	if actual != expected {
		t.Errorf("Expected '%s', received '%s'", expected, actual)
	}
}
