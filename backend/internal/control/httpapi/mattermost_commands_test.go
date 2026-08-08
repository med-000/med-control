package httpapi

import (
	"testing"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func TestParseCreateTaskText(t *testing.T) {
	command, err := parseCreateTaskText("write report 2026-08-10 09:30", nil)
	if err != nil {
		t.Fatal(err)
	}

	if command.Title != "write report" {
		t.Fatalf("Title = %q, want %q", command.Title, "write report")
	}
	if command.Date == nil || command.Date.Start == nil || command.Date.Start.Format("2006-01-02") != "2026-08-10" {
		t.Fatalf("Date = %+v", command.Date)
	}
	if command.Notification == nil || command.Notification.Start == nil || command.Notification.Start.Format("2006-01-02 15:04") != "2026-08-10 09:30" {
		t.Fatalf("Notification = %+v", command.Notification)
	}
}

func TestParseCreateTaskTextAppliesPriority(t *testing.T) {
	command, err := parseCreateTaskText("call back 2026-08-10 09:30", &taskdomain.SelectOption{Name: "High"})
	if err != nil {
		t.Fatal(err)
	}

	if command.Priority == nil || command.Priority.Name != "High" {
		t.Fatalf("Priority = %+v", command.Priority)
	}
}

func TestParseCreateTaskTextRejectsMissingFields(t *testing.T) {
	_, err := parseCreateTaskText("title only", nil)
	if err == nil {
		t.Fatal("error should be returned")
	}
}
