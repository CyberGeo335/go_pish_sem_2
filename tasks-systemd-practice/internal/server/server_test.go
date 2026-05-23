package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CyberGeo335/tasks/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter(config.Config{LogLevel: "silent"})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected health response to contain status ok, got %s", rec.Body.String())
	}
}

func TestCreateTask(t *testing.T) {
	router := NewRouter(config.Config{LogLevel: "silent"})
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"deploy app"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"title":"deploy app"`) {
		t.Fatalf("expected created task in response, got %s", rec.Body.String())
	}
}
