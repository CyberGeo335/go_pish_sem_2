package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer() *httptest.Server {
	store := NewStore()
	handler := NewHandler(store)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func TestHealthEndpoint(t *testing.T) {
	server := testServer()
	defer server.Close()

	response, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
}

func TestCreateAndGetTask(t *testing.T) {
	server := testServer()
	defer server.Close()

	body := bytes.NewBufferString(`{"title":"configure ci"}`)
	response, err := http.Post(server.URL+"/tasks", "application/json", body)
	if err != nil {
		t.Fatalf("POST /tasks failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.StatusCode)
	}

	var created Task
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode created task: %v", err)
	}
	if created.ID != 1 || created.Title != "configure ci" {
		t.Fatalf("unexpected task: %+v", created)
	}

	getResponse, err := http.Get(server.URL + "/tasks/1")
	if err != nil {
		t.Fatalf("GET /tasks/1 failed: %v", err)
	}
	defer getResponse.Body.Close()

	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResponse.StatusCode)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	server := testServer()
	defer server.Close()

	response, err := http.Post(server.URL+"/tasks", "application/json", strings.NewReader(`{"title":" "}`))
	if err != nil {
		t.Fatalf("POST /tasks failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}

func TestMarkDoneAndDeleteTask(t *testing.T) {
	server := testServer()
	defer server.Close()

	createResponse, err := http.Post(server.URL+"/tasks", "application/json", strings.NewReader(`{"title":"ship image"}`))
	if err != nil {
		t.Fatalf("POST /tasks failed: %v", err)
	}
	createResponse.Body.Close()

	request, err := http.NewRequest(http.MethodPatch, server.URL+"/tasks/1/done", nil)
	if err != nil {
		t.Fatalf("failed to create PATCH request: %v", err)
	}
	patchResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("PATCH /tasks/1/done failed: %v", err)
	}
	defer patchResponse.Body.Close()

	if patchResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", patchResponse.StatusCode)
	}

	deleteRequest, err := http.NewRequest(http.MethodDelete, server.URL+"/tasks/1", nil)
	if err != nil {
		t.Fatalf("failed to create DELETE request: %v", err)
	}
	deleteResponse, err := http.DefaultClient.Do(deleteRequest)
	if err != nil {
		t.Fatalf("DELETE /tasks/1 failed: %v", err)
	}
	defer deleteResponse.Body.Close()

	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", deleteResponse.StatusCode)
	}
}
