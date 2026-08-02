package task

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

var ErrTaskNotFound = errors.New("task not found")

type Repository interface {
	UpsertMany(ctx context.Context, tasks []taskdomain.Task) error
	FindByDisplayID(ctx context.Context, displayID string) (taskdomain.Task, error)
	ScheduleReminder(ctx context.Context, displayID string, remindAt time.Time) (taskdomain.Task, error)
	List(ctx context.Context, filter taskdomain.Filter) ([]taskdomain.Task, error)
	DueForNotification(ctx context.Context, now time.Time) ([]taskdomain.Task, error)
	MarkNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error
}

type Notifier interface {
	SendTaskNotification(ctx context.Context, task taskdomain.Task) error
}

type TaskCreator interface {
	CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error)
}

type Service struct {
	repository Repository
	notifier   Notifier
	creator    TaskCreator
}

func NewService(repository Repository, notifier Notifier, creator TaskCreator) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
		creator:    creator,
	}
}

func (service *Service) ImportTasks(ctx context.Context, tasks []taskdomain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	return service.repository.UpsertMany(ctx, tasks)
}

func (service *Service) ListTasks(ctx context.Context, filter taskdomain.Filter) ([]taskdomain.Task, error) {
	return service.repository.List(ctx, filter)
}

func (service *Service) CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error) {
	if service.creator == nil {
		return taskdomain.Task{}, fmt.Errorf("task creator is required")
	}
	if command.Title == "" {
		return taskdomain.Task{}, fmt.Errorf("task title is required")
	}

	task, err := service.creator.CreateTask(ctx, command)
	if err != nil {
		return taskdomain.Task{}, err
	}

	if err := service.repository.UpsertMany(ctx, []taskdomain.Task{task}); err != nil {
		return taskdomain.Task{}, err
	}

	return task, nil
}

func (service *Service) RemindTask(ctx context.Context, displayID string, after time.Duration, now time.Time) (taskdomain.Task, error) {
	if after <= 0 {
		return taskdomain.Task{}, fmt.Errorf("remind duration must be positive")
	}

	remindAt := now.Add(after)
	return service.repository.ScheduleReminder(ctx, displayID, remindAt)
}

func (service *Service) NotifyDueTasks(ctx context.Context, now time.Time) ([]taskdomain.Task, error) {
	if service.notifier == nil {
		return nil, fmt.Errorf("task notifier is required")
	}

	tasks, err := service.repository.DueForNotification(ctx, now)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if err := service.notifier.SendTaskNotification(ctx, task); err != nil {
			return nil, err
		}
		if err := service.repository.MarkNotificationSent(ctx, task, now); err != nil {
			return nil, err
		}
		log.Printf("task notification sent: task_id=%s display_id=%s title=%q", task.ID, task.DisplayID, task.Title)
	}

	return tasks, nil
}
