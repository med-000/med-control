package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	apptask "github.com/med-000/overview/backend/internal/app/task"
	taskdomain "github.com/med-000/overview/shared/domain/task"
)

func TestFindByDisplayIDMatchesNumberSuffix(t *testing.T) {
	repository := NewTaskRepository()
	task := taskdomain.Task{ID: "notion:page-id", DisplayID: "OVW-5", Title: "task"}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}

	got, err := repository.FindByDisplayID(context.Background(), "5")
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != task.ID {
		t.Fatalf("ID = %q, want %q", got.ID, task.ID)
	}
}

func TestFindByDisplayIDReturnsNotFound(t *testing.T) {
	repository := NewTaskRepository()

	_, err := repository.FindByDisplayID(context.Background(), "5")

	if !errors.Is(err, apptask.ErrTaskNotFound) {
		t.Fatalf("error = %v, want ErrTaskNotFound", err)
	}
}

func TestScheduleReminderSurvivesNotionSyncUpsert(t *testing.T) {
	repository := NewTaskRepository()
	originalNotification := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
	remindAt := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	task := taskdomain.Task{
		ID:        "notion:page-id",
		DisplayID: "5",
		Title:     "task",
		Notification: &taskdomain.DateRange{
			Start: &originalNotification,
		},
	}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}

	if _, err := repository.ScheduleReminder(context.Background(), "5", remindAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.UpsertMany(context.Background(), []taskdomain.Task{task}); err != nil {
		t.Fatal(err)
	}

	got, err := repository.FindByDisplayID(context.Background(), "5")
	if err != nil {
		t.Fatal(err)
	}
	if got.Notification == nil || got.Notification.Start == nil || !got.Notification.Start.Equal(remindAt) {
		t.Fatalf("Notification.Start = %v, want %v", got.Notification, remindAt)
	}
}
