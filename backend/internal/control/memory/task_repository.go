package memory

import (
	"context"
	"sync"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

type TaskRepository struct {
	mu                   sync.RWMutex
	tasks                map[string]taskdomain.Task
	sentNotificationKeys map[string]time.Time
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks:                make(map[string]taskdomain.Task),
		sentNotificationKeys: make(map[string]time.Time),
	}
}

func (repository *TaskRepository) UpsertMany(ctx context.Context, tasks []taskdomain.Task) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for _, task := range tasks {
		if task.ID == "" {
			continue
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
		if _, sent := repository.sentNotificationKeys[notificationKey(task)]; sent {
			continue
		}
		result = append(result, task)
	}
	return result, nil
}

func (repository *TaskRepository) MarkNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.sentNotificationKeys[notificationKey(task)] = sentAt
	return nil
}

func notificationKey(task taskdomain.Task) string {
	if task.Notification == nil || task.Notification.Start == nil {
		return task.ID
	}
	return task.ID + ":" + task.Notification.Start.Format(time.RFC3339Nano)
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
		if displayID == id {
			return true
		}
	}
	return false
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
