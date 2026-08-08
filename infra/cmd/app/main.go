package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/med-000/med-control/infra/internal/app/tasksync"
	"github.com/med-000/med-control/infra/internal/config"
	backendcontrol "github.com/med-000/med-control/infra/internal/control/backend"
	"github.com/med-000/med-control/infra/internal/control/httpapi"
	notioncontrol "github.com/med-000/med-control/infra/internal/control/notion"
)

func main() {
	cfg := config.Load()
	if cfg.NotionAPIKey == "" {
		exitWithError("NOTION_API_KEY is required")
	}
	if cfg.NotionDatabaseID == "" {
		exitWithError("NOTION_MED_CONTROL_DB_KEY is required")
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
	taskSource := notioncontrol.NewTaskRepository(notionClient, cfg.NotionDatabaseID)
	taskDestination := backendcontrol.NewTaskSender(cfg.BackendTasksEndpoint, cfg.HTTPTimeout)
	service := tasksync.NewService(taskSource, taskDestination)

	mux := http.NewServeMux()
	handler := httpapi.NewHandler(service, cfg.NotionWebhookVerificationToken)
	handler.Register(mux)

	server := &http.Server{
		Addr:    cfg.NotionWebhookAddr,
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("notion webhook server shutdown failed: %v", err)
		}
	}()
	go func() {
		log.Printf("notion webhook server listening on %s", cfg.NotionWebhookAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("notion webhook server failed: %v", err)
			stop()
		}
	}()

	log.Printf("notion sync worker started: interval=%s", cfg.NotionSyncInterval)
	service.RunSyncLoop(ctx, cfg.NotionSyncInterval)
}

func exitWithError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
