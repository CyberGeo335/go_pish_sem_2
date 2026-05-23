package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/task"
)

func newTestServer() *httptest.Server {
	repo := task.NewRepository()
	mux := http.NewServeMux()
	NewHandler(repo).Register(mux)
	return httptest.NewServer(mux)
}

func TestRESTListAndGet(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/tasks")
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(server.URL + "/v1/tasks/t_001")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRESTNotFound(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/tasks/unknown")
	if err != nil {
		t.Fatalf("get unknown task: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestRESTCreateAndPatch(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, err := http.Post(server.URL+"/v1/tasks", "application/json", strings.NewReader(`{"title":"Compare","description":"PZ12"}`))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodPatch, server.URL+"/v1/tasks/t_001", bytes.NewBufferString(`{"done":true}`))
	if err != nil {
		t.Fatalf("new patch request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("patch task: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
