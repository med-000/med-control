package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/med-000/overview/infra/internal/app/tasksync"
	"github.com/med-000/overview/infra/internal/config"
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

	notionClient := notioncontrol.NewClient(cfg.NotionAPIKey)
	taskSource := notioncontrol.NewTaskRepository(notionClient, cfg.NoitonOverviewDBKey)
	service := tasksync.NewService(taskSource, nil)

	tasks, err := service.FetchTasks(context.Background())
	if err != nil {
		exitWithError(err.Error())
	}

	output, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Println(string(output))
}

func exitWithError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
