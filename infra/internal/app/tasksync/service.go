package tasksync

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type TaskSource interface {
	FetchTasks(ctx context.Context) ([]taskdomain.Task, error)
	FetchTaskByPageID(ctx context.Context, pageID string) (taskdomain.Task, error)
	CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error)
	FetchRawRows(ctx context.Context) ([]map[string]any, error)
	ListTemplates(ctx context.Context) ([]Template, error)
	UpdateTaskStatus(ctx context.Context, pageID string, status string) (taskdomain.Task, error)
}

type TaskDestination interface {
	SendTasks(ctx context.Context, tasks []taskdomain.Task) error
}

type Template struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
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

func (service *Service) ListTemplates(ctx context.Context) ([]Template, error) {
	return service.source.ListTemplates(ctx)
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

func (service *Service) CompleteScheduleItems(ctx context.Context, now time.Time) ([]taskdomain.Task, error) {
	if service.destination == nil {
		return nil, fmt.Errorf("task destination is required")
	}

	tasks, err := service.source.FetchTasks(ctx)
	if err != nil {
		return nil, err
	}

	var updated []taskdomain.Task
	for _, task := range tasks {
		if !shouldCompleteScheduleItem(task, now) {
			continue
		}
		if task.SourceID == "" {
			continue
		}

		updatedTask, err := service.source.UpdateTaskStatus(ctx, task.SourceID, "done")
		if err != nil {
			return nil, err
		}
		updated = append(updated, updatedTask)
	}

	if len(updated) == 0 {
		return nil, nil
	}
	if err := service.destination.SendTasks(ctx, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func shouldCompleteScheduleItem(task taskdomain.Task, now time.Time) bool {
	if task.Label == nil || !strings.EqualFold(task.Label.Name, "schedule") {
		return false
	}
	if task.Status != nil && strings.EqualFold(task.Status.Name, "done") {
		return false
	}
	if task.Date == nil || task.Date.Start == nil {
		return false
	}

	localNow := now.In(time.Local)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.Local)
	start := task.Date.Start.In(time.Local)
	return start.Before(today)
}
