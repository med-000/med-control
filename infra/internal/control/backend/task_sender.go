package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

type TaskSender struct {
	endpoint   string
	httpClient *http.Client
}

func NewTaskSender(endpoint string) *TaskSender {
	return &TaskSender{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (sender *TaskSender) SendTasks(ctx context.Context, tasks []taskdomain.Task) error {
	body, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := sender.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("backend request failed: status=%d body=%s", response.StatusCode, string(responseBody))
	}

	return nil
}
