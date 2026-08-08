package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

type TaskCreator struct {
	endpoint   string
	httpClient *http.Client
}

func NewTaskCreator(endpoint string, timeout time.Duration) *TaskCreator {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &TaskCreator{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (creator *TaskCreator) CreateTask(ctx context.Context, command taskdomain.CreateCommand) (taskdomain.Task, error) {
	if creator.endpoint == "" {
		return taskdomain.Task{}, fmt.Errorf("INFRA_QUICK_TASK_ENDPOINT is required")
	}

	body, err := json.Marshal(map[string]string{
		"title": command.Title,
	})
	if err != nil {
		return taskdomain.Task{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, creator.endpoint, bytes.NewReader(body))
	if err != nil {
		return taskdomain.Task{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := creator.httpClient.Do(request)
	if err != nil {
		return taskdomain.Task{}, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return taskdomain.Task{}, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return taskdomain.Task{}, fmt.Errorf("infra request failed: status=%d body=%s", response.StatusCode, string(responseBody))
	}

	var task taskdomain.Task
	if err := json.Unmarshal(responseBody, &task); err != nil {
		return taskdomain.Task{}, err
	}

	return task, nil
}
