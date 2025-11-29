package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	cfg := &Config{
		Server:    ServerConfig{Port: ":8080"},
		JWTSecret: "dummy",
		Services:  []ServiceConfig{},
	}
	r := buildRouter(cfg)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rw := httptest.NewRecorder()

	r.ServeHTTP(rw, req)

	if got, want := rw.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status: got %d want %d", got, want)
	}
}
