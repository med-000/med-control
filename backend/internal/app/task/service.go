package task

import (
	"context"
	"fmt"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

type Repository interface {
	UpsertMany(ctx context.Context, tasks []taskdomain.Task) error
	List(ctx context.Context, filter taskdomain.Filter) ([]taskdomain.Task, error)
	DueForNotification(ctx context.Context, now time.Time) ([]taskdomain.Task, error)
	MarkNotificationSent(ctx context.Context, task taskdomain.Task, sentAt time.Time) error
}

type Notifier interface {
	SendTaskNotification(ctx context.Context, task taskdomain.Task) error
}

type Service struct {
	repository Repository
	notifier   Notifier
}

func NewService(repository Repository, notifier Notifier) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
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
	}

	return tasks, nil
}
