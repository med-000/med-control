package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	apptask "github.com/med-000/overview/backend/internal/app/task"
	"github.com/med-000/overview/backend/internal/config"
	"github.com/med-000/overview/backend/internal/control/httpapi"
	"github.com/med-000/overview/backend/internal/control/memory"
	"github.com/med-000/overview/infra/mattermost"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	taskRepository := memory.NewTaskRepository()
	taskNotifier := mattermost.NewNotifier(cfg.MattermostOverviewWebhook, cfg.HTTPTimeout)
	taskService := apptask.NewService(taskRepository, taskNotifier)
	go taskService.RunNotificationLoop(ctx, cfg.TaskNotifyInterval)

	mux := http.NewServeMux()
	handler := httpapi.NewHandler(taskService)
	handler.Register(mux)

	fmt.Printf("backend listening on %s\n", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
