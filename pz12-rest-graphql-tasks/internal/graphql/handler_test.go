package graphql

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/task"
)

func execute(t *testing.T, body string) map[string]any {
	t.Helper()
	repo := task.NewRepository()
	mux := http.NewServeMux()
	NewHandler(repo).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return payload
}

func TestGraphQLTasksSelectsRequestedFields(t *testing.T) {
	payload := execute(t, `{"query":"query { tasks { id title done } }"}`)
	data := payload["data"].(map[string]any)
	tasks := data["tasks"].([]any)
	first := tasks[0].(map[string]any)
	if _, ok := first["description"]; ok {
		t.Fatal("description must not be returned when it was not requested")
	}
	if _, ok := first["id"]; !ok {
		t.Fatal("id must be returned")
	}
}

func TestGraphQLTaskNotFoundReturnsErrorsField(t *testing.T) {
	payload := execute(t, `{"query":"query GetTask($id: ID!) { task(id: $id) { id title description done } }","variables":{"id":"unknown"}}`)
	if _, ok := payload["errors"]; !ok {
		t.Fatal("expected errors field")
	}
}

func TestGraphQLCreateTask(t *testing.T) {
	payload := execute(t, `{"query":"mutation Create($input: CreateTaskInput!) { createTask(input: $input) { id title description done } }","variables":{"input":{"title":"Compare REST and GraphQL","description":"PZ12"}}}`)
	data := payload["data"].(map[string]any)
	created := data["createTask"].(map[string]any)
	if created["id"] != "t_003" {
		t.Fatalf("expected id t_003, got %v", created["id"])
	}
}
