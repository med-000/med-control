package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	apptask "github.com/med-000/med-control/backend/internal/app/task"
	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type Handler struct {
	taskService *apptask.Service
	tokens      MattermostCommandTokens
	now         func() time.Time
}

type MattermostCommandTokens struct {
	Remind string
	Quick  string
	Create string
}

func NewHandler(taskService *apptask.Service, tokens MattermostCommandTokens) *Handler {
	return &Handler{
		taskService: taskService,
		tokens:      tokens,
		now:         time.Now,
	}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handler.health)
	mux.HandleFunc("POST /tasks/import", handler.importTasks)
	mux.HandleFunc("GET /tasks", handler.listTasks)
	mux.HandleFunc("POST /tasks/notify-due", handler.notifyDueTasks)
	mux.HandleFunc("POST /mattermost/commands/remind", handler.remindCommand)
	mux.HandleFunc("POST /mattermost/commands/quick", handler.quickCommand)
	mux.HandleFunc("POST /mattermost/commands/create", handler.createCommand)
}

func (handler *Handler) health(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (handler *Handler) importTasks(writer http.ResponseWriter, request *http.Request) {
	var tasks []taskdomain.Task
	if err := json.NewDecoder(request.Body).Decode(&tasks); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err := handler.taskService.ImportTasks(request.Context(), tasks); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(writer, http.StatusAccepted, map[string]any{
		"imported": len(tasks),
	})
}

func (handler *Handler) listTasks(writer http.ResponseWriter, request *http.Request) {
	tasks, err := handler.taskService.ListTasks(request.Context(), taskdomain.Filter{})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(writer, http.StatusOK, tasks)
}

func (handler *Handler) notifyDueTasks(writer http.ResponseWriter, request *http.Request) {
	tasks, err := handler.taskService.NotifyDueTasks(request.Context(), handler.now())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(writer, http.StatusAccepted, map[string]any{
		"notified": len(tasks),
	})
}

func (handler *Handler) validMattermostCommand(writer http.ResponseWriter, request *http.Request, token string) bool {
	if err := request.ParseForm(); err != nil {
		writeMattermostResponse(writer, http.StatusBadRequest, "request body is invalid")
		return false
	}

	if token == "" || request.FormValue("token") != token {
		http.Error(writer, "invalid mattermost command token", http.StatusUnauthorized)
		return false
	}

	return true
}

func writeMattermostResponse(writer http.ResponseWriter, status int, text string) {
	writeJSON(writer, status, map[string]string{
		"response_type": "in_channel",
		"text":          text,
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
