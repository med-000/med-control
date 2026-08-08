package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/med-000/med-control/infra/internal/app/tasksync"
	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type Handler struct {
	taskSync                 *tasksync.Service
	webhookVerificationToken string
}

func NewHandler(taskSync *tasksync.Service, webhookVerificationToken string) *Handler {
	return &Handler{
		taskSync:                 taskSync,
		webhookVerificationToken: webhookVerificationToken,
	}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handler.health)
	mux.HandleFunc("POST /notion/webhook", handler.notionWebhook)
	mux.HandleFunc("POST /tasks/quick", handler.quickTask)
}

func (handler *Handler) health(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (handler *Handler) quickTask(writer http.ResponseWriter, request *http.Request) {
	var command quickTaskRequest
	if err := json.NewDecoder(request.Body).Decode(&command); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(command.Title)
	if title == "" {
		http.Error(writer, "title is required", http.StatusBadRequest)
		return
	}

	task, err := handler.taskSync.CreateTask(request.Context(), taskdomain.CreateCommand{
		Title: title,
	})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(writer, http.StatusCreated, task)
}

func (handler *Handler) notionWebhook(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if handler.webhookVerificationToken != "" && !validNotionSignature(body, request.Header.Get("X-Notion-Signature"), handler.webhookVerificationToken) {
		http.Error(writer, "invalid notion webhook signature", http.StatusUnauthorized)
		return
	}

	var event notionWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if event.VerificationToken != "" {
		log.Printf("notion webhook verification token: %s", event.VerificationToken)
		writeJSON(writer, http.StatusOK, map[string]string{
			"status": "verification_token_received",
		})
		return
	}

	switch {
	case strings.HasPrefix(event.Type, "page.") && event.Entity.Type == "page" && event.Entity.ID != "":
		task, err := handler.taskSync.SyncTaskByPageID(request.Context(), event.Entity.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("notion webhook page sync completed: type=%s page_id=%s task_id=%s display_id=%s", event.Type, event.Entity.ID, task.ID, task.DisplayID)
		writeJSON(writer, http.StatusAccepted, map[string]any{
			"synced":     1,
			"event_type": event.Type,
			"page_id":    event.Entity.ID,
		})
	case shouldFullSync(event):
		tasks, err := handler.taskSync.SyncTasks(request.Context())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("notion webhook full sync completed: type=%s tasks=%d", event.Type, len(tasks))
		writeJSON(writer, http.StatusAccepted, map[string]any{
			"synced":     len(tasks),
			"event_type": event.Type,
		})
	default:
		log.Printf("notion webhook ignored: type=%s entity_type=%s entity_id=%s", event.Type, event.Entity.Type, event.Entity.ID)
		writeJSON(writer, http.StatusAccepted, map[string]any{
			"synced":     0,
			"event_type": event.Type,
		})
	}
}

func validNotionSignature(body []byte, signature string, verificationToken string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(verificationToken))
	_, _ = mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

func shouldFullSync(event notionWebhookEvent) bool {
	return strings.HasPrefix(event.Type, "database.") || strings.HasPrefix(event.Type, "data_source.")
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		log.Printf("write json response failed: %v", err)
	}
}

type notionWebhookEvent struct {
	VerificationToken string              `json:"verification_token"`
	ID                string              `json:"id"`
	Type              string              `json:"type"`
	Entity            notionWebhookEntity `json:"entity"`
}

type quickTaskRequest struct {
	Title string `json:"title"`
}

type notionWebhookEntity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (event notionWebhookEvent) String() string {
	return fmt.Sprintf("type=%s entity_type=%s entity_id=%s", event.Type, event.Entity.Type, event.Entity.ID)
}
