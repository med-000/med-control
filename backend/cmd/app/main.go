package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	apptask "github.com/med-000/med-control/backend/internal/app/task"
	"github.com/med-000/med-control/backend/internal/config"
	"github.com/med-000/med-control/backend/internal/control/httpapi"
	infracontrol "github.com/med-000/med-control/backend/internal/control/infra"
	sqlitecontrol "github.com/med-000/med-control/backend/internal/control/sqlite"
	"github.com/med-000/med-control/infra/mattermost"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	taskRepository, err := sqlitecontrol.NewTaskRepository(cfg.DBPath)
	if err != nil {
		exitWithError(fmt.Sprintf("open backend db failed: %v", err))
	}
	defer taskRepository.Close()

	taskNotifier := mattermost.NewNotifier(cfg.MattermostWebhook, cfg.HTTPTimeout)
	taskCreator := infracontrol.NewTaskCreator(cfg.InfraQuickTaskEndpoint, cfg.HTTPTimeout)
	taskService := apptask.NewService(taskRepository, taskNotifier, taskCreator)
	go taskService.RunNotificationLoop(ctx, cfg.TaskNotifyInterval)

	mux := http.NewServeMux()
	handler := httpapi.NewHandler(taskService, httpapi.MattermostCommandTokens{
		Remind: cfg.MattermostRemindToken,
		Quick:  cfg.MattermostQuickToken,
	})
	handler.Register(mux)

	fmt.Printf("backend listening on %s\n", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		exitWithError(err.Error())
	}
}

func exitWithError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
