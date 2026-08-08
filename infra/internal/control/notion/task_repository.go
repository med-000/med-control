package notion

import (
	"context"

	"github.com/med-000/med-control/infra/internal/app/tasksync"
	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type TaskRepository struct {
	client     *Client
	databaseID string
}

func NewTaskRepository(client *Client, databaseID string) *TaskRepository {
	return &TaskRepository{
		client:     client,
		databaseID: databaseID,
	}
}

func (repository *TaskRepository) FetchRawRows(ctx context.Context) ([]map[string]any, error) {
	return repository.client.RetrieveDatabaseRows(ctx, repository.databaseID)
}

func (repository *TaskRepository) ListTemplates(ctx context.Context) ([]tasksync.Template, error) {
	return repository.client.ListDatabaseTemplates(ctx, repository.databaseID)
}

func (repository *TaskRepository) FetchTasks(ctx context.Context) ([]taskdomain.Task, error) {
	rows, err := repository.client.RetrieveDatabaseTasks(ctx, repository.databaseID)
	if err != nil {
		return nil, err
	}

	return RowsToTasks(rows), nil
}

func (repository *TaskRepository) CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error) {
	row, err := repository.client.CreateDatabaseTask(ctx, repository.databaseID, command)
	if err != nil {
		return taskdomain.Task{}, err
	}

	return RowToTask(row), nil
}

func (repository *TaskRepository) UpdateTaskStatus(ctx context.Context, pageID string, status string) (taskdomain.Task, error) {
	row, err := repository.client.UpdatePageStatus(ctx, pageID, status)
	if err != nil {
		return taskdomain.Task{}, err
	}

	return RowToTask(row), nil
}

func (repository *TaskRepository) FetchTaskByPageID(ctx context.Context, pageID string) (taskdomain.Task, error) {
	row, err := repository.client.RetrievePageTask(ctx, pageID)
	if err != nil {
		return taskdomain.Task{}, err
	}

	return RowToTask(row), nil
}
