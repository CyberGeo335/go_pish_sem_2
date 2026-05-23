package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/CyberGeo335/pz12-rest-graphql-tasks/internal/task"
)

// Handler exposes the Task domain through a compact educational GraphQL endpoint.
// It supports the operations required by practical work #12: tasks, task,
// createTask, and updateTask. The project intentionally uses only Go standard
// library packages so it can run offline without downloading dependencies.
type Handler struct {
	repo *task.Repository
}

type request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type response struct {
	Data   map[string]any `json:"data,omitempty"`
	Errors []gqlError     `json:"errors,omitempty"`
}

type gqlError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

// NewHandler creates a GraphQL handler.
func NewHandler(repo *task.Repository) *Handler {
	return &Handler{repo: repo}
}

// Register attaches GraphQL routes to a mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/query", h.query)
	mux.HandleFunc("/graphql", h.playground)
}

func (h *Handler) query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{
			Errors: []gqlError{{Message: "GraphQL endpoint accepts POST requests only"}},
		})
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{
			Errors: []gqlError{{Message: "invalid json body"}},
		})
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		writeGraphQLError(w, "query", "query is required")
		return
	}

	switch {
	case containsField(query, "createTask"):
		h.createTask(w, req)
	case containsField(query, "updateTask"):
		h.updateTask(w, req)
	case containsField(query, "tasks"):
		h.tasks(w, req)
	case containsField(query, "task"):
		h.task(w, req)
	default:
		writeGraphQLError(w, "query", "unsupported operation")
	}
}

func (h *Handler) tasks(w http.ResponseWriter, req request) {
	fields := selectionFields(req.Query, "tasks")
	items := h.repo.List()

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, taskToMap(item, fields))
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{"tasks": result},
	})
}

func (h *Handler) task(w http.ResponseWriter, req request) {
	id := variableString(req.Variables, "id")
	if id == "" {
		id = argumentString(req.Query, "task", "id")
	}
	if id == "" {
		writeGraphQLError(w, "task", "id is required")
		return
	}

	item, err := h.repo.Get(id)
	if err != nil {
		writeTaskError(w, "task", err)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{"task": taskToMap(item, selectionFields(req.Query, "task"))},
	})
}

func (h *Handler) createTask(w http.ResponseWriter, req request) {
	input := variableMap(req.Variables, "input")
	if input == nil {
		writeGraphQLError(w, "createTask", "input is required")
		return
	}

	item, err := h.repo.Create(task.CreateInput{
		Title:       stringFromMap(input, "title"),
		Description: stringFromMap(input, "description"),
	})
	if err != nil {
		writeTaskError(w, "createTask", err)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{"createTask": taskToMap(item, selectionFields(req.Query, "createTask"))},
	})
}

func (h *Handler) updateTask(w http.ResponseWriter, req request) {
	id := variableString(req.Variables, "id")
	if id == "" {
		id = argumentString(req.Query, "updateTask", "id")
	}
	if id == "" {
		writeGraphQLError(w, "updateTask", "id is required")
		return
	}

	input := variableMap(req.Variables, "input")
	if input == nil {
		writeGraphQLError(w, "updateTask", "input is required")
		return
	}

	update := task.UpdateInput{}
	if v, ok := input["title"].(string); ok {
		update.Title = &v
	}
	if v, ok := input["description"].(string); ok {
		update.Description = &v
	}
	if v, ok := input["done"].(bool); ok {
		update.Done = &v
	}

	item, err := h.repo.Update(id, update)
	if err != nil {
		writeTaskError(w, "updateTask", err)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{"updateTask": taskToMap(item, selectionFields(req.Query, "updateTask"))},
	})
}

func writeTaskError(w http.ResponseWriter, path string, err error) {
	message := err.Error()
	if errors.Is(err, task.ErrNotFound) {
		message = "task not found"
	}
	writeJSON(w, http.StatusOK, response{
		Data:   map[string]any{path: nil},
		Errors: []gqlError{{Message: message, Path: []string{path}}},
	})
}

func writeGraphQLError(w http.ResponseWriter, path string, message string) {
	writeJSON(w, http.StatusOK, response{
		Data:   map[string]any{path: nil},
		Errors: []gqlError{{Message: message, Path: []string{path}}},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func taskToMap(item task.Task, fields []string) map[string]any {
	if len(fields) == 0 {
		fields = []string{"id", "title", "description", "done"}
	}

	result := make(map[string]any, len(fields))
	for _, field := range fields {
		switch field {
		case "id":
			result["id"] = item.ID
		case "title":
			result["title"] = item.Title
		case "description":
			result["description"] = item.Description
		case "done":
			result["done"] = item.Done
		}
	}
	return result
}

func variableString(vars map[string]any, key string) string {
	if vars == nil {
		return ""
	}
	value, _ := vars[key].(string)
	return value
}

func variableMap(vars map[string]any, key string) map[string]any {
	if vars == nil {
		return nil
	}
	if value, ok := vars[key].(map[string]any); ok {
		return value
	}
	return nil
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, _ := values[key].(string)
	return value
}

func containsField(query, field string) bool {
	for i := 0; i < len(query); i++ {
		if isFieldAt(query, i, field) {
			return true
		}
	}
	return false
}

func selectionFields(query, field string) []string {
	selection := fieldSelection(query, field)
	if selection == "" {
		return nil
	}

	allowed := map[string]struct{}{
		"id":          {},
		"title":       {},
		"description": {},
		"done":        {},
	}

	fields := make([]string, 0, 4)
	seen := make(map[string]struct{}, 4)
	for _, token := range identifierTokens(selection) {
		if _, ok := allowed[token]; !ok {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		fields = append(fields, token)
	}
	return fields
}

func fieldSelection(query, field string) string {
	for i := 0; i < len(query); i++ {
		if !isFieldAt(query, i, field) {
			continue
		}
		j := i + len(field)
		j = skipSpace(query, j)
		if j < len(query) && query[j] == '(' {
			end := matching(query, j, '(', ')')
			if end == -1 {
				return ""
			}
			j = end + 1
		}
		j = skipSpace(query, j)
		if j >= len(query) || query[j] != '{' {
			continue
		}
		end := matching(query, j, '{', '}')
		if end == -1 {
			return ""
		}
		return query[j+1 : end]
	}
	return ""
}

func argumentString(query, field, arg string) string {
	for i := 0; i < len(query); i++ {
		if !isFieldAt(query, i, field) {
			continue
		}
		j := skipSpace(query, i+len(field))
		if j >= len(query) || query[j] != '(' {
			return ""
		}
		end := matching(query, j, '(', ')')
		if end == -1 {
			return ""
		}
		args := query[j+1 : end]
		needle := arg + ":"
		idx := strings.Index(args, needle)
		if idx == -1 {
			return ""
		}
		value := strings.TrimSpace(args[idx+len(needle):])
		if comma := strings.Index(value, ","); comma != -1 {
			value = value[:comma]
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"`)
		if strings.HasPrefix(value, "$") {
			return ""
		}
		return value
	}
	return ""
}

func isFieldAt(query string, index int, field string) bool {
	if index < 0 || index+len(field) > len(query) || query[index:index+len(field)] != field {
		return false
	}
	beforeOK := index == 0 || !isIdentRune(rune(query[index-1]))
	afterIndex := index + len(field)
	afterOK := afterIndex >= len(query) || !isIdentRune(rune(query[afterIndex]))
	return beforeOK && afterOK
}

func skipSpace(s string, i int) int {
	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		i++
	}
	return i
}

func matching(s string, start int, open, close byte) int {
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == open {
			depth++
			continue
		}
		if c == close {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func identifierTokens(s string) []string {
	tokens := make([]string, 0)
	for i := 0; i < len(s); {
		r := rune(s[i])
		if !isIdentStart(r) {
			i++
			continue
		}
		start := i
		i++
		for i < len(s) && isIdentRune(rune(s[i])) {
			i++
		}
		tokens = append(tokens, s[start:i])
	}
	return tokens
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func (h *Handler) playground(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, response{
			Errors: []gqlError{{Message: "playground accepts GET requests only"}},
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <title>PZ12 GraphQL playground</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 32px; max-width: 1100px; }
    textarea { box-sizing: border-box; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; width: 100%; min-height: 150px; margin-bottom: 12px; }
    button { padding: 8px 14px; }
    pre { background: #f4f4f4; padding: 16px; overflow: auto; }
  </style>
</head>
<body>
  <h1>PZ12 GraphQL playground</h1>
  <p>Endpoint: <code>POST /query</code></p>
  <label>Query</label>
  <textarea id="query">query { tasks { id title done } }</textarea>
  <label>Variables JSON</label>
  <textarea id="variables">{}</textarea>
  <button onclick="send()">Execute</button>
  <pre id="result"></pre>
  <script>
    async function send() {
      const payload = {
        query: document.getElementById('query').value,
        variables: JSON.parse(document.getElementById('variables').value || '{}')
      };
      const response = await fetch('/query', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(payload)
      });
      document.getElementById('result').textContent = JSON.stringify(await response.json(), null, 2);
    }
  </script>
</body>
</html>`)
}
