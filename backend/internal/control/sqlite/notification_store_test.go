package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

func TestNotificationStorePersistsSentNotification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "overview.db")
	key := "notion:page-id:2026-08-05T09:00:00Z"

	store, err := NewNotificationStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkNotificationSent(context.Background(), key, taskdomain.Task{
		ID:        "notion:page-id",
		DisplayID: "5",
		Title:     "task",
	}, time.Date(2026, 8, 5, 9, 1, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	reopened, err := NewNotificationStore(path)
	if err != nil {
		t.Fatal(err)
	}

	sent, err := reopened.IsNotificationSent(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if !sent {
		t.Fatal("notification should be marked as sent after reopening store")
	}
}
