package main

import (
	"context"
	"fmt"
	"os"

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

	notionClient := notioncontrol.NewClient(cfg.NotionAPIKey)
	taskSource := notioncontrol.NewTaskRepository(notionClient, cfg.NoitonOverviewDBKey)
	taskDestination := backendcontrol.NewTaskSender(cfg.BackendTasksEndpoint)
	service := tasksync.NewService(taskSource, taskDestination)

	tasks, err := service.SyncTasks(context.Background())
	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Printf("sent %d tasks to backend\n", len(tasks))
}

func exitWithError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
