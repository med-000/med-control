package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	apptask "github.com/med-000/overview/backend/internal/app/task"
	taskdomain "github.com/med-000/overview/shared/domain/task"
)

type Handler struct {
	taskService *apptask.Service
	now         func() time.Time
}

func NewHandler(taskService *apptask.Service) *Handler {
	return &Handler{
		taskService: taskService,
		now:         time.Now,
	}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /tasks/import", handler.importTasks)
	mux.HandleFunc("GET /tasks", handler.listTasks)
	mux.HandleFunc("POST /tasks/notify-due", handler.notifyDueTasks)
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

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
