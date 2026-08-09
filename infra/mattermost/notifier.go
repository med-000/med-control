package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
	"github.com/med-000/med-control/shared/timeutil"
)

type Notifier struct {
	webhookURL string
	httpClient *http.Client
}

func NewNotifier(webhookURL string, timeout time.Duration) *Notifier {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Notifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (notifier *Notifier) SendTaskNotification(ctx context.Context, task taskdomain.Task) error {
	if notifier.webhookURL == "" {
		return fmt.Errorf("mattermost webhook URL is required")
	}

	body, err := json.Marshal(map[string]string{
		"text": formatTaskMessage(task),
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, notifier.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := notifier.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("mattermost request failed: status=%d body=%s", response.StatusCode, string(responseBody))
	}

	return nil
}

func formatTaskMessage(task taskdomain.Task) string {
	var lines []string
	lines = append(lines, "### "+task.DisplayTitle())

	if task.Date != nil && task.Date.Start != nil {
		lines = append(lines, "- date: "+formatDateTime(*task.Date.Start, task.Date.TimeZone))
	}
	if task.Status != nil {
		lines = append(lines, "- status: "+task.Status.Name)
	}
	if task.Priority != nil {
		lines = append(lines, "- priority: "+task.Priority.Name)
	}
	if task.Label != nil {
		lines = append(lines, "- label: "+task.Label.Name)
	}
	if len(task.Categories) > 0 {
		lines = append(lines, "- categories: "+joinOptionNames(task.Categories))
	}
	if task.SourceURL != "" {
		lines = append(lines, task.SourceURL)
	}
	if task.Description != "" {
		lines = append(lines, "", task.Description)
	}

	return strings.Join(lines, "\n")
}

func formatDateTime(value time.Time, timeZone string) string {
	location := timeutil.JST()
	if timeZone != "" {
		if loaded, err := time.LoadLocation(timeZone); err == nil {
			location = loaded
		}
	}

	return value.In(location).Format("2006-01-02 15:04")
}

func joinOptionNames(options []taskdomain.SelectOption) string {
	names := make([]string, 0, len(options))
	for _, option := range options {
		names = append(names, option.Name)
	}
	return strings.Join(names, ", ")
}
