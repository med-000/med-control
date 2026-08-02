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
	MattermostOverviewWebhook string
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
		MattermostOverviewWebhook: os.Getenv("MATTERMOST_OVERVIEW_WEBHOOK"),
		TaskNotifyInterval:        durationFromSeconds(os.Getenv("TASK_NOTIFY_INTERVAL_SECONDS"), 0),
		HTTPTimeout:               durationFromSeconds(os.Getenv("HTTP_TIMEOUT_SECONDS"), 10*time.Second),
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
