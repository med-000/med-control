package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr                      string
	DBPath                    string
	MattermostOverviewWebhook string
	MattermostCommandToken    string
	MattermostRemindToken     string
	MattermostQuickToken      string
	InfraQuickTaskEndpoint    string
	TaskNotifyInterval        time.Duration
	HTTPTimeout               time.Duration
}

func Load() Config {
	_ = loadEnvFile(".env")
	_ = loadEnvFile("backend/.env")
	_ = loadEnvFile("../../.env")

	addr := os.Getenv("BACKEND_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{
		Addr:                      addr,
		DBPath:                    stringWithFallback(os.Getenv("BACKEND_DB_PATH"), "/data/overview.db"),
		MattermostOverviewWebhook: os.Getenv("MATTERMOST_OVERVIEW_WEBHOOK"),
		MattermostCommandToken:    os.Getenv("MATTERMOST_COMMAND_TOKEN"),
		MattermostRemindToken:     firstNonEmpty(os.Getenv("MATTERMOST_REMIND_COMMAND_TOKEN"), os.Getenv("MATTERMOST_COMMAND_TOKEN")),
		MattermostQuickToken:      firstNonEmpty(os.Getenv("MATTERMOST_QUICK_COMMAND_TOKEN"), os.Getenv("MATTERMOST_COMMAND_TOKEN")),
		InfraQuickTaskEndpoint:    os.Getenv("INFRA_QUICK_TASK_ENDPOINT"),
		TaskNotifyInterval:        durationFromSeconds(os.Getenv("TASK_NOTIFY_INTERVAL_SECONDS"), 0),
		HTTPTimeout:               durationFromSeconds(os.Getenv("HTTP_TIMEOUT_SECONDS"), 10*time.Second),
	}
}

func stringWithFallback(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func durationFromSeconds(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
