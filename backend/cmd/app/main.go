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
	infracontrol "github.com/med-000/overview/backend/internal/control/infra"
	"github.com/med-000/overview/backend/internal/control/memory"
	sqlitecontrol "github.com/med-000/overview/backend/internal/control/sqlite"
	"github.com/med-000/overview/infra/mattermost"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	notificationStore, err := sqlitecontrol.NewNotificationStore(cfg.DBPath)
	if err != nil {
		exitWithError(fmt.Sprintf("open backend db failed: %v", err))
	}
	defer notificationStore.Close()

	taskRepository := memory.NewTaskRepository(notificationStore)
	taskNotifier := mattermost.NewNotifier(cfg.MattermostOverviewWebhook, cfg.HTTPTimeout)
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
