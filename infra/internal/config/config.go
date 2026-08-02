package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type Config struct {
	NotionAPIKey         string
	NoitonOverviewDBKey  string
	BackendTasksEndpoint string
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
	}
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
