package sqlite

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func TestItemRepositoryPersistsItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "med-control.db")
	repository := newTestItemRepository(t, path)

	notificationAt := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	task := taskdomain.Task{
		ID:        "notion:page-id",
		DisplayID: "TASK-7",
		Title:     "idea that behaves like task",
		Status:    &taskdomain.SelectOption{ID: "status-id", Name: "Doing", Color: "blue"},
		Categories: []taskdomain.SelectOption{
			{ID: "cat-id", Name: "Idea", Color: "green"},
		},
		Notification: &taskdomain.DateRange{Start: &notificationAt},
		Source:       taskdomain.SourceNotion,
		SourceID:     "page-id",
		SourceURL:    "https://notion.so/page-id",
		Description:  "body",
	}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}

	reopened := newTestItemRepository(t, path)
	tasks, err := reopened.List(context.Background(), taskdomain.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	got := tasks[0]
	if got.ID != task.ID || got.DisplayID != task.DisplayID || got.Title != task.Title {
		t.Fatalf("task = %+v, want %+v", got, task)
	}
	if len(got.Categories) != 1 || got.Categories[0].Name != "Idea" {
		t.Fatalf("categories = %+v", got.Categories)
	}
	if got.Notification == nil || got.Notification.Start == nil || !got.Notification.Start.Equal(notificationAt) {
		t.Fatalf("notification = %+v", got.Notification)
	}
}

func TestItemRepositoryPersistsLargeNotionBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "med-control.db")
	repository := newTestItemRepository(t, path)

	description := strings.Repeat("large notion body ", 20000)
	task := taskdomain.Task{
		ID:          "notion:large-page-id",
		DisplayID:   "TASK-8",
		Title:       "large body",
		Source:      taskdomain.SourceNotion,
		SourceID:    "large-page-id",
		Description: description,
	}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}

	got, err := repository.FindByDisplayID(context.Background(), "8")
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != description {
		t.Fatalf("description length = %d, want %d", len(got.Description), len(description))
	}
}

func TestItemRepositoryPersistsReminderOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "med-control.db")
	repository := newTestItemRepository(t, path)

	originalNotification := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	task := taskdomain.Task{
		ID:           "notion:page-id",
		DisplayID:    "TASK-7",
		Title:        "task",
		Notification: &taskdomain.DateRange{Start: &originalNotification},
		Source:       taskdomain.SourceNotion,
		SourceID:     "page-id",
	}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}

	override := time.Date(2026, 8, 9, 11, 0, 0, 0, time.UTC)
	if _, err := repository.ScheduleReminder(context.Background(), "7", override); err != nil {
		t.Fatal(err)
	}

	reopened := newTestItemRepository(t, path)
	got, err := reopened.FindByDisplayID(context.Background(), "7")
	if err != nil {
		t.Fatal(err)
	}
	if got.Notification == nil || got.Notification.Start == nil || !got.Notification.Start.Equal(override) {
		t.Fatalf("notification override = %+v, want %s", got.Notification, override)
	}
}

func TestItemRepositorySkipsSentNotificationsAndClearsOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "med-control.db")
	repository := newTestItemRepository(t, path)

	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	task := taskdomain.Task{
		ID:        "notion:page-id",
		DisplayID: "TASK-7",
		Title:     "task",
		Source:    taskdomain.SourceNotion,
		SourceID:  "page-id",
	}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.ScheduleReminder(context.Background(), "7", now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}

	due, err := repository.DueForNotification(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(due) != 1 {
		t.Fatalf("len(due) = %d, want 1", len(due))
	}
	if err := repository.MarkNotificationSent(context.Background(), due[0], now); err != nil {
		t.Fatal(err)
	}

	reopened := newTestItemRepository(t, path)
	due, err = reopened.DueForNotification(context.Background(), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(due) != 0 {
		t.Fatalf("len(due) = %d, want 0", len(due))
	}

	got, err := reopened.FindByDisplayID(context.Background(), "7")
	if err != nil {
		t.Fatal(err)
	}
	if got.Notification != nil {
		t.Fatalf("notification override should be cleared after sent: %+v", got.Notification)
	}
}

func newTestItemRepository(t *testing.T, path string) *ItemRepository {
	t.Helper()
	repository, err := NewItemRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}
