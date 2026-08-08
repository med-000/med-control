package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	apptask "github.com/med-000/med-control/backend/internal/app/task"
	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func (handler *Handler) remindCommand(writer http.ResponseWriter, request *http.Request) {
	if !handler.validMattermostCommand(writer, request, handler.tokens.Remind) {
		return
	}

	displayID, minutes, err := parseRemindText(request.FormValue("text"))
	if err != nil {
		writeMattermostResponse(writer, http.StatusOK, "usage: /remind <No> <minutes>")
		return
	}

	task, err := handler.taskService.RemindTask(request.Context(), displayID, time.Duration(minutes)*time.Minute, handler.now())
	if err != nil {
		if errors.Is(err, apptask.ErrTaskNotFound) {
			writeMattermostResponse(writer, http.StatusOK, fmt.Sprintf("No=%s の task が見つかりません", displayID))
			return
		}
		writeMattermostResponse(writer, http.StatusOK, "remind の登録に失敗しました: "+err.Error())
		return
	}

	writeMattermostResponse(writer, http.StatusOK, fmt.Sprintf("%s を %d 分後に再通知します", task.DisplayTitle(), minutes))
}

func (handler *Handler) createCommand(writer http.ResponseWriter, request *http.Request) {
	handler.createTaskFromMattermostCommand(writer, request, handler.tokens.Create, "create", nil)
}

func (handler *Handler) quickCommand(writer http.ResponseWriter, request *http.Request) {
	handler.createTaskFromMattermostCommand(writer, request, handler.tokens.Quick, "quick", &taskdomain.SelectOption{
		Name:  "High",
		Color: "red",
	})
}

func (handler *Handler) workCommand(writer http.ResponseWriter, request *http.Request) {
	if !handler.validMattermostCommand(writer, request, handler.tokens.Work) {
		return
	}
	if handler.workTemplateID == "" {
		writeMattermostResponse(writer, http.StatusOK, "NOTION_WORK_TEMPLATE_ID is required")
		return
	}

	command, err := parseWorkText(request.FormValue("text"), handler.now(), handler.workTemplateID)
	if err != nil {
		writeMattermostResponse(writer, http.StatusOK, "usage: /work <start|end|todo> <start_mmdd> <end_mmdd>")
		return
	}

	task, err := handler.taskService.CreateTask(request.Context(), command)
	if err != nil {
		writeMattermostResponse(writer, http.StatusOK, "work task の作成に失敗しました: "+err.Error())
		return
	}

	writeMattermostResponse(writer, http.StatusOK, "作成しました: "+task.DisplayTitle())
}

func (handler *Handler) createTaskFromMattermostCommand(writer http.ResponseWriter, request *http.Request, token string, commandName string, priority *taskdomain.SelectOption) {
	if !handler.validMattermostCommand(writer, request, token) {
		return
	}

	command, err := parseCreateTaskText(request.FormValue("text"), priority)
	if err != nil {
		writeMattermostResponse(writer, http.StatusOK, fmt.Sprintf("usage: /%s <title> <date> <notification>", commandName))
		return
	}

	task, err := handler.taskService.CreateTask(request.Context(), command)
	if err != nil {
		writeMattermostResponse(writer, http.StatusOK, commandName+" task の作成に失敗しました: "+err.Error())
		return
	}

	writeMattermostResponse(writer, http.StatusOK, "作成しました: "+task.DisplayTitle())
}

func parseWorkText(text string, now time.Time, templateID string) (taskdomain.CreateCommand, error) {
	fields := strings.Fields(text)
	if len(fields) != 3 {
		return taskdomain.CreateCommand{}, fmt.Errorf("invalid work text")
	}

	status, err := workStatus(fields[0])
	if err != nil {
		return taskdomain.CreateCommand{}, err
	}
	start, err := parseMMDD(fields[1], now)
	if err != nil {
		return taskdomain.CreateCommand{}, err
	}
	end, err := parseMMDD(fields[2], now)
	if err != nil {
		return taskdomain.CreateCommand{}, err
	}
	if end.Before(start) {
		return taskdomain.CreateCommand{}, fmt.Errorf("end date must not be before start date")
	}

	return taskdomain.CreateCommand{
		Title:      "ollo勤務",
		Status:     status,
		Date:       &taskdomain.DateRange{Start: &start, End: &end},
		TemplateID: templateID,
	}, nil
}

func workStatus(value string) (*taskdomain.SelectOption, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "start":
		return &taskdomain.SelectOption{Name: "inprogress"}, nil
	case "end":
		return &taskdomain.SelectOption{Name: "done"}, nil
	case "todo":
		return &taskdomain.SelectOption{Name: "todo"}, nil
	default:
		return nil, fmt.Errorf("invalid work status: %s", value)
	}
}

func parseMMDD(value string, now time.Time) (time.Time, error) {
	if len(value) != 4 {
		return time.Time{}, fmt.Errorf("invalid mmdd: %s", value)
	}
	month, err := strconv.Atoi(value[:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month: %s", value[:2])
	}
	day, err := strconv.Atoi(value[2:])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day: %s", value[2:])
	}

	result := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local)
	if result.Month() != time.Month(month) || result.Day() != day {
		return time.Time{}, fmt.Errorf("invalid mmdd: %s", value)
	}
	return result, nil
}

func parseRemindText(text string) (string, int, error) {
	fields := strings.Fields(text)
	if len(fields) != 2 {
		return "", 0, fmt.Errorf("invalid remind text")
	}

	minutes, err := strconv.Atoi(fields[1])
	if err != nil || minutes <= 0 {
		return "", 0, fmt.Errorf("invalid remind minutes")
	}

	return fields[0], minutes, nil
}

func parseCreateTaskText(text string, priority *taskdomain.SelectOption) (taskdomain.CreateCommand, error) {
	fields := strings.Fields(text)
	if len(fields) < 3 {
		return taskdomain.CreateCommand{}, fmt.Errorf("invalid create task text")
	}

	date, err := parseCommandDate(fields[len(fields)-2])
	if err != nil {
		return taskdomain.CreateCommand{}, err
	}
	notification, err := parseCommandNotification(fields[len(fields)-1], date)
	if err != nil {
		return taskdomain.CreateCommand{}, err
	}

	title := strings.TrimSpace(strings.Join(fields[:len(fields)-2], " "))
	if title == "" {
		return taskdomain.CreateCommand{}, fmt.Errorf("title is required")
	}

	return taskdomain.CreateCommand{
		Title:        title,
		Date:         &taskdomain.DateRange{Start: &date},
		Notification: &taskdomain.DateRange{Start: &notification},
		Priority:     cloneSelectOption(priority),
	}, nil
}

func parseCommandDate(value string) (time.Time, error) {
	for _, layout := range []string{time.DateOnly, time.RFC3339, "2006-01-02T15:04"} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", value)
}

func parseCommandNotification(value string, date time.Time) (time.Time, error) {
	if parsed, err := time.ParseInLocation("15:04", value, time.Local); err == nil {
		return time.Date(date.Year(), date.Month(), date.Day(), parsed.Hour(), parsed.Minute(), 0, 0, time.Local), nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", time.DateOnly} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid notification: %s", value)
}

func cloneSelectOption(option *taskdomain.SelectOption) *taskdomain.SelectOption {
	if option == nil {
		return nil
	}
	clone := *option
	return &clone
}
