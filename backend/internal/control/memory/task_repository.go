package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	apptask "github.com/med-000/med-control/backend/internal/app/task"
	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type TaskRepository struct {
	mu                   sync.RWMutex
	tasks                map[string]taskdomain.Task
	reminderOverrides    map[string]taskdomain.DateRange
	sentNotificationKeys map[string]time.Time
	notificationStore    NotificationStore
}

type NotificationStore interface {
	IsNotificationSent(ctx context.Context, notificationKey string) (bool, error)
	MarkNotificationSent(ctx context.Context, notificationKey string, task taskdomain.Task, sentAt time.Time) error
}

func NewTaskRepository(stores ...NotificationStore) *TaskRepository {
	var notificationStore NotificationStore
	if len(stores) > 0 {
		notificationStore = stores[0]
	}

	return &TaskRepository{
		tasks:                make(map[string]taskdomain.Task),
		reminderOverrides:    make(map[string]taskdomain.DateRange),
		sentNotificationKeys: make(map[string]time.Time),
		notificationStore:    notificationStore,
	}
}

func (repository *TaskRepository) UpsertMany(ctx context.Context, tasks []taskdomain.Task) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for _, task := range tasks {
		if task.ID == "" {
			continue
		}
		if override, exists := repository.reminderOverrides[task.ID]; exists {
			task.Notification = &override
		}
		repository.tasks[task.ID] = task
	}
	return nil
}

func (repository *TaskRepository) List(ctx context.Context, filter taskdomain.Filter) ([]taskdomain.Task, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	result := make([]taskdomain.Task, 0, len(repository.tasks))
	for _, task := range repository.tasks {
		if !matchesFilter(task, filter) {
			continue
		}
		result = append(result, task)
	}
	return result, nil
}

func (repository *TaskRepository) FindByDisplayID(ctx context.Context, displayID string) (taskdomain.Task, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	for _, task := range repository.tasks {
		if displayIDMatches(task.DisplayID, displayID) {
			return task, nil
		}
	}
	return taskdomain.Task{}, apptask.ErrTaskNotFound
}

func (repository *TaskRepository) ScheduleReminder(ctx context.Context, displayID string, remindAt time.Time) (taskdomain.Task, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for id, task := range repository.tasks {
		if !displayIDMatches(task.DisplayID, displayID) {
			continue
		}

		override := taskdomain.DateRange{Start: &remindAt}
		task.Notification = &override
		repository.tasks[id] = task
		repository.reminderOverrides[id] = override
		return task, nil
	}
	return taskdomain.Task{}, apptask.ErrTaskNotFound
}

func (repository *TaskRepository) DueForNotification(ctx context.Context, now time.Time) ([]taskdomain.Task, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	var result []taskdomain.Task
	for _, task := range repository.tasks {
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

func (repository *TaskRepository) MarkNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	if err := repository.markNotificationSent(ctx, task, sentAt); err != nil {
		return err
	}
	if override, exists := repository.reminderOverrides[task.ID]; exists && sameNotificationStart(&override, task.Notification) {
		delete(repository.reminderOverrides, task.ID)
	}
	return nil
}

func (repository *TaskRepository) isNotificationSent(ctx context.Context, task taskdomain.Task) (bool, error) {
	key := notificationKey(task)
	if repository.notificationStore != nil {
		return repository.notificationStore.IsNotificationSent(ctx, key)
	}
	_, sent := repository.sentNotificationKeys[key]
	return sent, nil
}

func (repository *TaskRepository) markNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error {
	key := notificationKey(task)
	if repository.notificationStore != nil {
		return repository.notificationStore.MarkNotificationSent(ctx, key, task, sentAt)
	}
	repository.sentNotificationKeys[key] = sentAt
	return nil
}

func notificationKey(task taskdomain.Task) string {
	if task.Notification == nil || task.Notification.Start == nil {
		return task.ID
	}
	return task.ID + ":" + task.Notification.Start.Format(time.RFC3339Nano)
}

func sameNotificationStart(left *taskdomain.DateRange, right *taskdomain.DateRange) bool {
	if left == nil || right == nil || left.Start == nil || right.Start == nil {
		return false
	}
	return left.Start.Equal(*right.Start)
}

func matchesFilter(task taskdomain.Task, filter taskdomain.Filter) bool {
	if len(filter.DisplayIDs) > 0 && !displayIDIn(task.DisplayID, filter.DisplayIDs) {
		return false
	}
	if len(filter.StatusIDs) > 0 && !selectIDIn(task.Status, filter.StatusIDs) {
		return false
	}
	if len(filter.StatusNames) > 0 && !selectNameIn(task.Status, filter.StatusNames) {
		return false
	}
	if len(filter.LabelIDs) > 0 && !selectIDIn(task.Label, filter.LabelIDs) {
		return false
	}
	if len(filter.LabelNames) > 0 && !selectNameIn(task.Label, filter.LabelNames) {
		return false
	}
	if len(filter.PriorityIDs) > 0 && !selectIDIn(task.Priority, filter.PriorityIDs) {
		return false
	}
	if len(filter.CategoryIDs) > 0 && !selectIDsInclude(task.Categories, filter.CategoryIDs) {
		return false
	}
	if filter.From != nil && (task.Date == nil || task.Date.Start == nil || task.Date.Start.Before(*filter.From)) {
		return false
	}
	if filter.To != nil && (task.Date == nil || task.Date.Start == nil || task.Date.Start.After(*filter.To)) {
		return false
	}
	return true
}

func displayIDIn(displayID string, ids []string) bool {
	for _, id := range ids {
		if displayIDMatches(displayID, id) {
			return true
		}
	}
	return false
}

func displayIDMatches(displayID string, query string) bool {
	displayID = strings.ToLower(strings.TrimSpace(displayID))
	query = strings.ToLower(strings.TrimSpace(query))
	if displayID == "" || query == "" {
		return false
	}
	if displayID == query {
		return true
	}

	hyphen := strings.LastIndex(displayID, "-")
	return hyphen >= 0 && displayID[hyphen+1:] == query
}

func selectIDIn(option *taskdomain.SelectOption, ids []string) bool {
	if option == nil {
		return false
	}
	for _, id := range ids {
		if option.ID == id {
			return true
		}
	}
	return false
}

func selectNameIn(option *taskdomain.SelectOption, names []string) bool {
	if option == nil {
		return false
	}
	for _, name := range names {
		if option.Name == name {
			return true
		}
	}
	return false
}

func selectIDsInclude(options []taskdomain.SelectOption, ids []string) bool {
	for _, option := range options {
		for _, id := range ids {
			if option.ID == id {
				return true
			}
		}
	}
	return false
}
