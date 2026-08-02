package sqlite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

type NotificationStore struct {
	mu   sync.RWMutex
	path string
	sent map[string]struct{}
}

func NewNotificationStore(path string) (*NotificationStore, error) {
	if path == "" {
		path = "/data/overview.db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	store := &NotificationStore{
		path: path,
		sent: make(map[string]struct{}),
	}
	if err := store.migrate(context.Background()); err != nil {
		return nil, err
	}
	if err := store.loadSentKeys(context.Background()); err != nil {
		return nil, err
	}

	return store, nil
}

func (store *NotificationStore) Close() error {
	return nil
}

func (store *NotificationStore) IsNotificationSent(ctx context.Context, notificationKey string) (bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	_, sent := store.sent[notificationKey]
	return sent, nil
}

func (store *NotificationStore) MarkNotificationSent(ctx context.Context, notificationKey string, task taskdomain.Task, sentAt time.Time) error {
	var notificationAt string
	if task.Notification != nil && task.Notification.Start != nil {
		notificationAt = task.Notification.Start.Format(time.RFC3339Nano)
	}

	sql := fmt.Sprintf(
		`INSERT OR IGNORE INTO sent_notifications (
			notification_key,
			task_id,
			display_id,
			title,
			notification_at,
			sent_at
		) VALUES (%s, %s, %s, %s, %s, %s);`,
		sqlString(notificationKey),
		sqlString(task.ID),
		sqlString(task.DisplayID),
		sqlString(task.Title),
		sqlNullableString(notificationAt),
		sqlString(sentAt.Format(time.RFC3339Nano)),
	)
	if err := store.exec(ctx, sql); err != nil {
		return err
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.sent[notificationKey] = struct{}{}
	return nil
}

func (store *NotificationStore) migrate(ctx context.Context) error {
	return store.exec(ctx, `
CREATE TABLE IF NOT EXISTS sent_notifications (
	notification_key TEXT PRIMARY KEY,
	task_id TEXT NOT NULL,
	display_id TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	notification_at TEXT,
	sent_at TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sent_notifications_task_id
	ON sent_notifications(task_id);

CREATE INDEX IF NOT EXISTS idx_sent_notifications_sent_at
	ON sent_notifications(sent_at);
`)
}

func (store *NotificationStore) loadSentKeys(ctx context.Context) error {
	output, err := store.query(ctx, `SELECT notification_key FROM sent_notifications;`)
	if err != nil {
		return err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key != "" {
			store.sent[key] = struct{}{}
		}
	}
	return scanner.Err()
}

func (store *NotificationStore) exec(ctx context.Context, sql string) error {
	_, err := store.query(ctx, sql)
	return err
}

func (store *NotificationStore) query(ctx context.Context, sql string) (string, error) {
	command := exec.CommandContext(ctx, "sqlite3", store.path, sql)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sqlite3 failed: %w output=%s", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func sqlNullableString(value string) string {
	if value == "" {
		return "NULL"
	}
	return sqlString(value)
}
