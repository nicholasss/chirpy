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

func TestHandlerReady(t *testing.T) {
	expected := "OK"

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	handlerReady(w, req)

	res := w.Result()
	defer res.Body.Close()
	bodyData, err := io.ReadAll(res.Body)
	actual := string(bodyData)

	if err != nil {
		t.Errorf("Error was found with handlerReady: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected '%d', recieved '%d'", http.StatusOK, res.StatusCode)
	}
	if actual != expected {
		t.Errorf("Expected '%s', received '%s'", expected, actual)
	}
}
