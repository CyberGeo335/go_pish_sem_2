package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/CyberGeo335/pz14-task-queue/internal/jobs"
	"github.com/CyberGeo335/pz14-task-queue/internal/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultAuthToken = "demo-token"

type Handler struct {
	channel *amqp.Channel
	mu      sync.Mutex
}

func NewHandler(channel *amqp.Channel) *Handler {
	return &Handler{channel: channel}
}

type createJobRequest struct {
	TaskID string `json:"task_id"`
}

type createJobResponse struct {
	Status    string `json:"status"`
	TaskID    string `json:"task_id"`
	MessageID string `json:"message_id"`
	Queue     string `json:"queue"`
}

func (h *Handler) ProcessTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !validBearerToken(r) {
		writeError(w, http.StatusUnauthorized, "missing or invalid Authorization header")
		return
	}

	var req createJobRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.TaskID = strings.TrimSpace(req.TaskID)
	if req.TaskID == "" {
		writeError(w, http.StatusBadRequest, "task_id is required")
		return
	}

	messageID, err := newUUIDv4()
	if err != nil {
		log.Printf("message id generation error: %v", err)
		writeError(w, http.StatusInternalServerError, "message id generation error")
		return
	}

	job := jobs.TaskJob{
		Job:       "process_task",
		TaskID:    req.TaskID,
		Attempt:   1,
		MessageID: messageID,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// amqp091-go channels should not be published to concurrently.
	h.mu.Lock()
	err = rabbitmq.PublishJSON(ctx, h.channel, rabbitmq.TaskJobsQueue, job, messageID)
	h.mu.Unlock()
	if err != nil {
		log.Printf("publish job error: %v", err)
		writeError(w, http.StatusInternalServerError, "could not publish job")
		return
	}

	log.Printf("accepted task_id=%s message_id=%s queue=%s", job.TaskID, job.MessageID, rabbitmq.TaskJobsQueue)

	writeJSON(w, http.StatusAccepted, createJobResponse{
		Status:    "accepted",
		TaskID:    job.TaskID,
		MessageID: job.MessageID,
		Queue:     rabbitmq.TaskJobsQueue,
	})
}

func Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func validBearerToken(r *http.Request) bool {
	token := os.Getenv("AUTH_TOKEN")
	if token == "" {
		token = defaultAuthToken
	}
	return r.Header.Get("Authorization") == "Bearer "+token
}

func newUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
