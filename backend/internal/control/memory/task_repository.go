package memory

import (
	"context"
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
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks:                make(map[string]taskdomain.Task),
		reminderOverrides:    make(map[string]taskdomain.DateRange),
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
		if !taskdomain.MatchesFilter(task, filter) {
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
		if taskdomain.DisplayIDMatches(task.DisplayID, displayID) {
			return task, nil
		}
	}
	return taskdomain.Task{}, apptask.ErrTaskNotFound
}

func (repository *TaskRepository) ScheduleReminder(ctx context.Context, displayID string, remindAt time.Time) (taskdomain.Task, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for id, task := range repository.tasks {
		if !taskdomain.DisplayIDMatches(task.DisplayID, displayID) {
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
	if override, exists := repository.reminderOverrides[task.ID]; exists && taskdomain.SameNotificationStart(&override, task.Notification) {
		delete(repository.reminderOverrides, task.ID)
	}
	return nil
}

func (repository *TaskRepository) isNotificationSent(ctx context.Context, task taskdomain.Task) (bool, error) {
	key := taskdomain.NotificationKey(task)
	_, sent := repository.sentNotificationKeys[key]
	return sent, nil
}

func (repository *TaskRepository) markNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error {
	key := taskdomain.NotificationKey(task)
	repository.sentNotificationKeys[key] = sentAt
	return nil
}
