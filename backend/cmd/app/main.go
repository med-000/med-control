package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apptask "github.com/med-000/med-control/backend/internal/app/task"
	"github.com/med-000/med-control/backend/internal/config"
	"github.com/med-000/med-control/backend/internal/control/httpapi"
	infracontrol "github.com/med-000/med-control/backend/internal/control/infra"
	sqlitecontrol "github.com/med-000/med-control/backend/internal/control/sqlite"
	"github.com/med-000/med-control/infra/mattermost"
	"github.com/med-000/med-control/shared/timeutil"
)

func main() {
	time.Local = timeutil.JST()

	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	taskRepository, err := sqlitecontrol.NewItemRepository(cfg.DBPath)
	if err != nil {
		exitWithError(fmt.Sprintf("open backend db failed: %v", err))
	}
	defer taskRepository.Close()

	taskNotifier, err := newTaskNotifier(cfg)
	if err != nil {
		exitWithError(err.Error())
	}
	taskCreator := infracontrol.NewTaskCreator(cfg.InfraQuickTaskEndpoint, cfg.HTTPTimeout)
	taskService := apptask.NewService(taskRepository, taskNotifier, taskCreator)
	go taskService.RunNotificationLoop(ctx, cfg.TaskNotifyInterval)

	mux := http.NewServeMux()
	handler := httpapi.NewHandler(taskService, httpapi.MattermostCommandTokens{
		Remind: cfg.MattermostRemindToken,
		Quick:  cfg.MattermostQuickToken,
		Create: cfg.MattermostCreateToken,
		Work:   cfg.MattermostWorkToken,
	}, cfg.NotionWorkTemplateID)
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

func newTaskNotifier(cfg config.Config) (apptask.Notifier, error) {
	botValues := []string{
		cfg.MattermostBotAPIURL,
		cfg.MattermostBotToken,
		cfg.MattermostBotChannelID,
	}
	botConfiguredCount := 0
	for _, value := range botValues {
		if value != "" {
			botConfiguredCount++
		}
	}

	if botConfiguredCount == len(botValues) {
		return mattermost.NewBotNotifier(cfg.MattermostBotAPIURL, cfg.MattermostBotToken, cfg.MattermostBotChannelID, cfg.HTTPTimeout), nil
	}
	if botConfiguredCount > 0 {
		return nil, fmt.Errorf("MATTERMOST_BOT_API_URL, MATTERMOST_BOT_TOKEN, and MATTERMOST_BOT_CHANNEL_ID must all be set to use Mattermost bot notifications")
	}

	return mattermost.NewNotifier(cfg.MattermostWebhook, cfg.HTTPTimeout), nil
}
