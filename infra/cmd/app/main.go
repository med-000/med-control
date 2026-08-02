package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/med-000/overview/infra/internal/app/tasksync"
	"github.com/med-000/overview/infra/internal/config"
	backendcontrol "github.com/med-000/overview/infra/internal/control/backend"
	notioncontrol "github.com/med-000/overview/infra/internal/control/notion"
)

func main() {
	cfg := config.Load()
	if cfg.NotionAPIKey == "" {
		exitWithError("NOTION_API_KEY is required")
	}
	if cfg.NoitonOverviewDBKey == "" {
		exitWithError("NOITON_OVERVIEW_DB_KEY is required")
	}
	if cfg.BackendTasksEndpoint == "" {
		exitWithError("BACKEND_TASKS_ENDPOINT is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	notionClient := notioncontrol.NewClient(notioncontrol.ClientConfig{
		APIKey:     cfg.NotionAPIKey,
		BaseURL:    cfg.NotionAPIBaseURL,
		APIVersion: cfg.NotionAPIVersion,
		PageSize:   cfg.NotionPageSize,
		Timeout:    cfg.HTTPTimeout,
	})
	taskSource := notioncontrol.NewTaskRepository(notionClient, cfg.NoitonOverviewDBKey)
	taskDestination := backendcontrol.NewTaskSender(cfg.BackendTasksEndpoint, cfg.HTTPTimeout)
	service := tasksync.NewService(taskSource, taskDestination)

	log.Printf("notion sync worker started: interval=%s", cfg.NotionSyncInterval)
	service.RunSyncLoop(ctx, cfg.NotionSyncInterval)
}

func exitWithError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
