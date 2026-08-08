package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	apptask "github.com/med-000/med-control/backend/internal/app/task"
	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type ItemRepository struct {
	mu   sync.RWMutex
	path string
}

func NewItemRepository(path string) (*ItemRepository, error) {
	if path == "" {
		path = "/data/med-control.db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	repository := &ItemRepository{path: path}
	if err := repository.migrate(context.Background()); err != nil {
		return nil, err
	}
	return repository, nil
}

func (repository *ItemRepository) Close() error {
	return nil
}

func (repository *ItemRepository) UpsertMany(ctx context.Context, tasks []taskdomain.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	var sql strings.Builder
	sql.WriteString("BEGIN;\n")
	for _, task := range tasks {
		pageID := notionPageID(task)
		if pageID == "" {
			continue
		}
		sql.WriteString(fmt.Sprintf(`
INSERT INTO items (
	notion_page_id,
	item_id,
	display_id,
	title,
	status_id,
	status_name,
	status_color,
	date_start,
	date_end,
	date_time_zone,
	label_id,
	label_name,
	label_color,
	priority_id,
	priority_name,
	priority_color,
	notification_start,
	notification_end,
	notification_time_zone,
	source_url,
	description,
	synced_at,
	last_seen_at,
	deleted_at
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NULL)
ON CONFLICT(notion_page_id) DO UPDATE SET
	item_id = excluded.item_id,
	display_id = excluded.display_id,
	title = excluded.title,
	status_id = excluded.status_id,
	status_name = excluded.status_name,
	status_color = excluded.status_color,
	date_start = excluded.date_start,
	date_end = excluded.date_end,
	date_time_zone = excluded.date_time_zone,
	label_id = excluded.label_id,
	label_name = excluded.label_name,
	label_color = excluded.label_color,
	priority_id = excluded.priority_id,
	priority_name = excluded.priority_name,
	priority_color = excluded.priority_color,
	notification_start = excluded.notification_start,
	notification_end = excluded.notification_end,
	notification_time_zone = excluded.notification_time_zone,
	source_url = excluded.source_url,
	description = excluded.description,
	synced_at = excluded.synced_at,
	last_seen_at = excluded.last_seen_at,
	deleted_at = NULL;
`,
			sqlString(pageID),
			sqlString(task.ID),
			sqlString(task.DisplayID),
			sqlString(task.Title),
			sqlNullableString(selectID(task.Status)),
			sqlNullableString(selectName(task.Status)),
			sqlNullableString(selectColor(task.Status)),
			sqlNullableString(dateStart(task.Date)),
			sqlNullableString(dateEnd(task.Date)),
			sqlNullableString(dateTimeZone(task.Date)),
			sqlNullableString(selectID(task.Label)),
			sqlNullableString(selectName(task.Label)),
			sqlNullableString(selectColor(task.Label)),
			sqlNullableString(selectID(task.Priority)),
			sqlNullableString(selectName(task.Priority)),
			sqlNullableString(selectColor(task.Priority)),
			sqlNullableString(dateStart(task.Notification)),
			sqlNullableString(dateEnd(task.Notification)),
			sqlNullableString(dateTimeZone(task.Notification)),
			sqlString(task.SourceURL),
			sqlString(task.Description),
			sqlString(now),
			sqlString(now),
		))

		sql.WriteString(fmt.Sprintf("DELETE FROM item_categories WHERE notion_page_id = %s;\n", sqlString(pageID)))
		for index, category := range task.Categories {
			sql.WriteString(fmt.Sprintf(
				`INSERT INTO item_categories (notion_page_id, category_id, category_name, category_color, position) VALUES (%s, %s, %s, %s, %d);`+"\n",
				sqlString(pageID),
				sqlString(category.ID),
				sqlString(category.Name),
				sqlString(category.Color),
				index,
			))
		}
	}
	sql.WriteString("COMMIT;\n")

	return repository.exec(ctx, sql.String())
}

func (repository *ItemRepository) List(ctx context.Context, filter taskdomain.Filter) ([]taskdomain.Task, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	tasks, err := repository.list(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]taskdomain.Task, 0, len(tasks))
	for _, task := range tasks {
		if !taskdomain.MatchesFilter(task, filter) {
			continue
		}
		result = append(result, task)
	}
	return result, nil
}

func (repository *ItemRepository) FindByDisplayID(ctx context.Context, displayID string) (taskdomain.Task, error) {
	tasks, err := repository.List(ctx, taskdomain.Filter{})
	if err != nil {
		return taskdomain.Task{}, err
	}
	for _, task := range tasks {
		if taskdomain.DisplayIDMatches(task.DisplayID, displayID) {
			return task, nil
		}
	}
	return taskdomain.Task{}, apptask.ErrTaskNotFound
}

func (repository *ItemRepository) ScheduleReminder(ctx context.Context, displayID string, remindAt time.Time) (taskdomain.Task, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	task, err := repository.findByDisplayIDLocked(ctx, displayID)
	if err != nil {
		return taskdomain.Task{}, err
	}

	pageID := notionPageID(task)
	override := taskdomain.DateRange{Start: &remindAt}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := repository.exec(ctx, fmt.Sprintf(`
INSERT INTO item_control_state (
	notion_page_id,
	reminder_override_start,
	reminder_override_end,
	reminder_override_time_zone,
	reminder_updated_at,
	reminder_source
) VALUES (%s, %s, NULL, '', %s, 'mattermost')
ON CONFLICT(notion_page_id) DO UPDATE SET
	reminder_override_start = excluded.reminder_override_start,
	reminder_override_end = excluded.reminder_override_end,
	reminder_override_time_zone = excluded.reminder_override_time_zone,
	reminder_updated_at = excluded.reminder_updated_at,
	reminder_source = excluded.reminder_source;
`,
		sqlString(pageID),
		sqlString(remindAt.UTC().Format(time.RFC3339Nano)),
		sqlString(now),
	)); err != nil {
		return taskdomain.Task{}, err
	}

	task.Notification = &override
	return task, nil
}

func (repository *ItemRepository) DueForNotification(ctx context.Context, now time.Time) ([]taskdomain.Task, error) {
	tasks, err := repository.List(ctx, taskdomain.Filter{})
	if err != nil {
		return nil, err
	}

	var result []taskdomain.Task
	for _, task := range tasks {
		if task.Notification == nil || task.Notification.Start == nil {
			continue
		}
		if task.Notification.Start.After(now) {
			continue
		}
		sent, err := repository.isNotificationSent(ctx, task)
		if err != nil {
			return nil, err
		}
		if sent {
			continue
		}
		result = append(result, task)
	}
	return result, nil
}

func (repository *ItemRepository) MarkNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	key := taskdomain.NotificationKey(task)
	pageID := notionPageID(task)
	var notificationAt string
	if task.Notification != nil && task.Notification.Start != nil {
		notificationAt = task.Notification.Start.Format(time.RFC3339Nano)
	}

	sql := fmt.Sprintf(`
BEGIN;
INSERT OR IGNORE INTO sent_notifications (
	notification_key,
	notion_page_id,
	task_id,
	display_id,
	title,
	notification_at,
	sent_at
) VALUES (%s, %s, %s, %s, %s, %s, %s);
DELETE FROM item_control_state
WHERE notion_page_id = %s
  AND reminder_override_start = %s;
COMMIT;
`,
		sqlString(key),
		sqlNullableString(pageID),
		sqlString(task.ID),
		sqlString(task.DisplayID),
		sqlString(task.Title),
		sqlNullableString(notificationAt),
		sqlString(sentAt.UTC().Format(time.RFC3339Nano)),
		sqlString(pageID),
		sqlNullableString(notificationAt),
	)
	return repository.exec(ctx, sql)
}

func (repository *ItemRepository) migrate(ctx context.Context) error {
	if err := repository.exec(ctx, `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS items (
	notion_page_id TEXT PRIMARY KEY,
	item_id TEXT NOT NULL UNIQUE,
	display_id TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL,
	status_id TEXT,
	status_name TEXT,
	status_color TEXT,
	date_start TEXT,
	date_end TEXT,
	date_time_zone TEXT,
	label_id TEXT,
	label_name TEXT,
	label_color TEXT,
	priority_id TEXT,
	priority_name TEXT,
	priority_color TEXT,
	notification_start TEXT,
	notification_end TEXT,
	notification_time_zone TEXT,
	source_url TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	synced_at TEXT NOT NULL,
	last_seen_at TEXT NOT NULL,
	deleted_at TEXT
);

CREATE TABLE IF NOT EXISTS item_categories (
	notion_page_id TEXT NOT NULL,
	category_id TEXT NOT NULL DEFAULT '',
	category_name TEXT NOT NULL,
	category_color TEXT NOT NULL DEFAULT '',
	position INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (notion_page_id, category_id, category_name),
	FOREIGN KEY (notion_page_id) REFERENCES items(notion_page_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS item_control_state (
	notion_page_id TEXT PRIMARY KEY,
	reminder_override_start TEXT,
	reminder_override_end TEXT,
	reminder_override_time_zone TEXT,
	reminder_updated_at TEXT,
	reminder_source TEXT NOT NULL DEFAULT '',
	FOREIGN KEY (notion_page_id) REFERENCES items(notion_page_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sent_notifications (
	notification_key TEXT PRIMARY KEY,
	notion_page_id TEXT,
	task_id TEXT NOT NULL,
	display_id TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	notification_at TEXT,
	sent_at TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (notion_page_id) REFERENCES items(notion_page_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_items_display_id
	ON items(display_id);

CREATE INDEX IF NOT EXISTS idx_items_deleted_at
	ON items(deleted_at);

CREATE INDEX IF NOT EXISTS idx_sent_notifications_task_id
	ON sent_notifications(task_id);

CREATE INDEX IF NOT EXISTS idx_sent_notifications_sent_at
	ON sent_notifications(sent_at);
`); err != nil {
		return err
	}
	return repository.ensureSentNotificationsNotionPageID(ctx)
}

func (repository *ItemRepository) ensureSentNotificationsNotionPageID(ctx context.Context) error {
	output, err := repository.query(ctx, `PRAGMA table_info(sent_notifications);`)
	if err != nil {
		return err
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		return err
	}
	for _, row := range rows {
		if stringFromAny(row["name"]) == "notion_page_id" {
			return nil
		}
	}
	return repository.exec(ctx, `ALTER TABLE sent_notifications ADD COLUMN notion_page_id TEXT;`)
}

func (repository *ItemRepository) list(ctx context.Context) ([]taskdomain.Task, error) {
	output, err := repository.query(ctx, `
SELECT
	items.*,
	item_control_state.reminder_override_start,
	item_control_state.reminder_override_end,
	item_control_state.reminder_override_time_zone
FROM items
LEFT JOIN item_control_state ON item_control_state.notion_page_id = items.notion_page_id
WHERE items.deleted_at IS NULL
ORDER BY items.display_id, items.title;
`)
	if err != nil {
		return nil, err
	}

	var rows []itemRow
	if strings.TrimSpace(output) != "" {
		if err := json.Unmarshal([]byte(output), &rows); err != nil {
			return nil, err
		}
	}

	categories, err := repository.categories(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make([]taskdomain.Task, 0, len(rows))
	for _, row := range rows {
		task := row.task()
		task.Categories = categories[row.NotionPageID]
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (repository *ItemRepository) categories(ctx context.Context) (map[string][]taskdomain.SelectOption, error) {
	output, err := repository.query(ctx, `
SELECT notion_page_id, category_id, category_name, category_color
FROM item_categories
ORDER BY notion_page_id, position;
`)
	if err != nil {
		return nil, err
	}

	var rows []categoryRow
	if strings.TrimSpace(output) != "" {
		if err := json.Unmarshal([]byte(output), &rows); err != nil {
			return nil, err
		}
	}

	result := make(map[string][]taskdomain.SelectOption)
	for _, row := range rows {
		result[row.NotionPageID] = append(result[row.NotionPageID], taskdomain.SelectOption{
			ID:    row.CategoryID,
			Name:  row.CategoryName,
			Color: row.CategoryColor,
		})
	}
	return result, nil
}

func (repository *ItemRepository) findByDisplayIDLocked(ctx context.Context, displayID string) (taskdomain.Task, error) {
	tasks, err := repository.list(ctx)
	if err != nil {
		return taskdomain.Task{}, err
	}
	for _, task := range tasks {
		if taskdomain.DisplayIDMatches(task.DisplayID, displayID) {
			return task, nil
		}
	}
	return taskdomain.Task{}, apptask.ErrTaskNotFound
}

func (repository *ItemRepository) isNotificationSent(ctx context.Context, task taskdomain.Task) (bool, error) {
	output, err := repository.query(ctx, fmt.Sprintf(
		`SELECT notification_key FROM sent_notifications WHERE notification_key = %s LIMIT 1;`,
		sqlString(taskdomain.NotificationKey(task)),
	))
	if err != nil {
		return false, err
	}
	output = strings.TrimSpace(output)
	return output != "" && output != "[]", nil
}

func (repository *ItemRepository) exec(ctx context.Context, sql string) error {
	command := exec.CommandContext(ctx, "sqlite3", "-bail", repository.path, sql)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sqlite3 failed: %w output=%s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (repository *ItemRepository) query(ctx context.Context, sql string) (string, error) {
	command := exec.CommandContext(ctx, "sqlite3", "-json", repository.path, sql)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sqlite3 failed: %w output=%s", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

type itemRow struct {
	NotionPageID             string `json:"notion_page_id"`
	ItemID                   string `json:"item_id"`
	DisplayID                string `json:"display_id"`
	Title                    string `json:"title"`
	StatusID                 string `json:"status_id"`
	StatusName               string `json:"status_name"`
	StatusColor              string `json:"status_color"`
	DateStart                string `json:"date_start"`
	DateEnd                  string `json:"date_end"`
	DateTimeZone             string `json:"date_time_zone"`
	LabelID                  string `json:"label_id"`
	LabelName                string `json:"label_name"`
	LabelColor               string `json:"label_color"`
	PriorityID               string `json:"priority_id"`
	PriorityName             string `json:"priority_name"`
	PriorityColor            string `json:"priority_color"`
	NotificationStart        string `json:"notification_start"`
	NotificationEnd          string `json:"notification_end"`
	NotificationTimeZone     string `json:"notification_time_zone"`
	SourceURL                string `json:"source_url"`
	Description              string `json:"description"`
	ReminderOverrideStart    string `json:"reminder_override_start"`
	ReminderOverrideEnd      string `json:"reminder_override_end"`
	ReminderOverrideTimeZone string `json:"reminder_override_time_zone"`
}

func (row itemRow) task() taskdomain.Task {
	notification := dateRange(row.NotificationStart, row.NotificationEnd, row.NotificationTimeZone)
	if override := dateRange(row.ReminderOverrideStart, row.ReminderOverrideEnd, row.ReminderOverrideTimeZone); override != nil {
		notification = override
	}
	return taskdomain.Task{
		ID:           row.ItemID,
		DisplayID:    row.DisplayID,
		Title:        row.Title,
		Status:       selectOption(row.StatusID, row.StatusName, row.StatusColor),
		Date:         dateRange(row.DateStart, row.DateEnd, row.DateTimeZone),
		Label:        selectOption(row.LabelID, row.LabelName, row.LabelColor),
		Priority:     selectOption(row.PriorityID, row.PriorityName, row.PriorityColor),
		Notification: notification,
		Source:       taskdomain.SourceNotion,
		SourceID:     row.NotionPageID,
		SourceURL:    row.SourceURL,
		Description:  row.Description,
	}
}

type categoryRow struct {
	NotionPageID  string `json:"notion_page_id"`
	CategoryID    string `json:"category_id"`
	CategoryName  string `json:"category_name"`
	CategoryColor string `json:"category_color"`
}

func notionPageID(task taskdomain.Task) string {
	if task.SourceID != "" {
		return task.SourceID
	}
	return strings.TrimPrefix(task.ID, string(taskdomain.SourceNotion)+":")
}

func selectOption(id string, name string, color string) *taskdomain.SelectOption {
	if name == "" {
		return nil
	}
	return &taskdomain.SelectOption{ID: id, Name: name, Color: color}
}

func selectID(option *taskdomain.SelectOption) string {
	if option == nil {
		return ""
	}
	return option.ID
}

func selectName(option *taskdomain.SelectOption) string {
	if option == nil {
		return ""
	}
	return option.Name
}

func selectColor(option *taskdomain.SelectOption) string {
	if option == nil {
		return ""
	}
	return option.Color
}

func dateRange(start string, end string, timeZone string) *taskdomain.DateRange {
	startTime := parseTime(start)
	endTime := parseTime(end)
	if startTime == nil && endTime == nil {
		return nil
	}
	return &taskdomain.DateRange{Start: startTime, End: endTime, TimeZone: timeZone}
}

func dateStart(value *taskdomain.DateRange) string {
	if value == nil || value.Start == nil {
		return ""
	}
	return value.Start.UTC().Format(time.RFC3339Nano)
}

func dateEnd(value *taskdomain.DateRange) string {
	if value == nil || value.End == nil {
		return ""
	}
	return value.End.UTC().Format(time.RFC3339Nano)
}

func dateTimeZone(value *taskdomain.DateRange) string {
	if value == nil {
		return ""
	}
	return value.TimeZone
}

func parseTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func stringFromAny(value any) string {
	result, _ := value.(string)
	return result
}
