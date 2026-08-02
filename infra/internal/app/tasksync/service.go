package tasksync

import (
	"context"
	"fmt"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

type TaskSource interface {
	FetchTasks(ctx context.Context) ([]taskdomain.Task, error)
	FetchTaskByPageID(ctx context.Context, pageID string) (taskdomain.Task, error)
	CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error)
	FetchRawRows(ctx context.Context) ([]map[string]any, error)
}

type TaskDestination interface {
	SendTasks(ctx context.Context, tasks []taskdomain.Task) error
}

type Service struct {
	source      TaskSource
	destination TaskDestination
}

func NewService(source TaskSource, destination TaskDestination) *Service {
	return &Service{
		source:      source,
		destination: destination,
	}
}

func (service *Service) FetchTasks(ctx context.Context) ([]taskdomain.Task, error) {
	return service.source.FetchTasks(ctx)
}

func (service *Service) FetchRawRows(ctx context.Context) ([]map[string]any, error) {
	return service.source.FetchRawRows(ctx)
}

func (service *Service) SyncTasks(ctx context.Context) ([]taskdomain.Task, error) {
	if service.destination == nil {
		return nil, fmt.Errorf("task destination is required")
	}

	tasks, err := service.source.FetchTasks(ctx)
	if err != nil {
		return nil, err
	}

	if err := service.destination.SendTasks(ctx, tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (service *Service) SyncTaskByPageID(ctx context.Context, pageID string) (taskdomain.Task, error) {
	if service.destination == nil {
		return taskdomain.Task{}, fmt.Errorf("task destination is required")
	}

	task, err := service.source.FetchTaskByPageID(ctx, pageID)
	if err != nil {
		return taskdomain.Task{}, err
	}

	if err := service.destination.SendTasks(ctx, []taskdomain.Task{task}); err != nil {
		return taskdomain.Task{}, err
	}

	return task, nil
}

func (service *Service) CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error) {
	if service.destination == nil {
		return taskdomain.Task{}, fmt.Errorf("task destination is required")
	}

	task, err := service.source.CreateTask(ctx, command)
	if err != nil {
		return taskdomain.Task{}, err
	}

	if err := service.destination.SendTasks(ctx, []taskdomain.Task{task}); err != nil {
		return taskdomain.Task{}, err
	}

	return task, nil
}
