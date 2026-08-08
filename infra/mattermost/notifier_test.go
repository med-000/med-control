package mattermost

import (
	"strings"
	"testing"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

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
