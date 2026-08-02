package notion

import (
	"context"

	taskdomain "github.com/med-000/overview/shared/domain/task"
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

func (repository *TaskRepository) FetchTasks(ctx context.Context) ([]taskdomain.Task, error) {
	rows, err := repository.client.RetrieveDatabaseTasks(ctx, repository.databaseID)
	if err != nil {
		return nil, err
	}

	return RowsToTasks(rows), nil
}
