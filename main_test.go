package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCensorString(t *testing.T) {
	var tests = []struct {
		send string
		want string
	}{
		{"words", "words"},
		{"no censoring", "no censoring"},
		{"kerfuffle", "****"},
		{"sharbert", "****"},
		{"fornax", "****"},
		{"this is a long sentence", "this is a long sentence"},
	}

	for _, test := range tests {
		want := censorString(test.send)
		if want != test.want {
			t.Errorf("Expected '%s', want '%s'", test.want, want)
		}
	}
}

func TestHandlerReady(t *testing.T) {
	wantBody := "OK"

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	handlerReady(w, req)

	res := w.Result()
	defer res.Body.Close()
	bodyData, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error was found with handlerReady: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected '%d', recieved '%d'", http.StatusOK, res.StatusCode)
	}
	if string(bodyData) != wantBody {
		t.Errorf("Expected '%s', received '%s'", wantBody, string(bodyData))
	}
}
