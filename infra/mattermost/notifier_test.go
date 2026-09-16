package mattermost

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func TestBotNotifierPostsTaskMessage(t *testing.T) {
	task := taskdomain.Task{
		ID:        "notion:page-id",
		Title:     "task",
		Source:    taskdomain.SourceNotion,
		SourceID:  "page-id",
		SourceURL: "https://notion.so/page-id",
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v4/posts" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer bot-token" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}

		var payload map[string]string
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["channel_id"] != "channel-id" {
			t.Fatalf("channel_id = %q", payload["channel_id"])
		}
		if !strings.Contains(payload["message"], "### task") {
			t.Fatalf("message = %q", payload["message"])
		}

		writer.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	notifier := NewBotNotifier(server.URL, "bot-token", "channel-id", time.Second)
	if err := notifier.SendTaskNotification(t.Context(), task); err != nil {
		t.Fatalf("SendTaskNotification: %v", err)
	}
}

func TestFormatTaskMessageShowsDateInJSTByDefault(t *testing.T) {
	start := time.Date(2026, 8, 10, 0, 30, 0, 0, time.UTC)
	task := taskdomain.Task{
		ID:        "notion:page-id",
		Title:     "task",
		Date:      &taskdomain.DateRange{Start: &start},
		Source:    taskdomain.SourceNotion,
		SourceID:  "page-id",
		SourceURL: "https://notion.so/page-id",
	}

	message := formatTaskMessage(task)
	if !strings.Contains(message, "- date: 2026-08-10 09:30") {
		t.Fatalf("message = %q", message)
	}
}

func TestFormatTaskMessageUsesDateTimeZone(t *testing.T) {
	start := time.Date(2026, 8, 10, 0, 30, 0, 0, time.UTC)
	task := taskdomain.Task{
		ID:       "notion:page-id",
		Title:    "task",
		Date:     &taskdomain.DateRange{Start: &start, TimeZone: "America/New_York"},
		Source:   taskdomain.SourceNotion,
		SourceID: "page-id",
	}

	message := formatTaskMessage(task)
	if !strings.Contains(message, "- date: 2026-08-09 20:30") {
		t.Fatalf("message = %q", message)
	}
}
