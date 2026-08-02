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
	NotionAPIKey         string
	NoitonOverviewDBKey  string
	BackendTasksEndpoint string
	NotionSyncInterval   time.Duration
	NotionAPIBaseURL     string
	NotionAPIVersion     string
	NotionPageSize       int
	HTTPTimeout          time.Duration
}

func Load() Config {
	_ = loadEnvFile(".env")
	_ = loadEnvFile("infra/.env")
	_ = loadEnvFile("../../.env")

	return Config{
		NotionAPIKey: os.Getenv("NOTION_API_KEY"),
		NoitonOverviewDBKey: firstNonEmpty(
			os.Getenv("NOITON_OVERVIEW_DB_KEY"),
			os.Getenv("NOTION_OVERVIEW_DB_KEY"),
		),
		BackendTasksEndpoint: os.Getenv("BACKEND_TASKS_ENDPOINT"),
		NotionSyncInterval:   durationFromSeconds(os.Getenv("NOTION_SYNC_INTERVAL_SECONDS"), 5*time.Minute),
		NotionAPIBaseURL:     stringWithFallback(os.Getenv("NOTION_API_BASE_URL"), "https://api.notion.com"),
		NotionAPIVersion:     stringWithFallback(os.Getenv("NOTION_API_VERSION"), "2026-03-11"),
		NotionPageSize:       intWithFallback(os.Getenv("NOTION_PAGE_SIZE"), 100),
		HTTPTimeout:          durationFromSeconds(os.Getenv("HTTP_TIMEOUT_SECONDS"), 10*time.Second),
	}
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

func intWithFallback(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil || result <= 0 {
		return fallback
	}
	return result
}

func stringWithFallback(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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
