package httpapi

import (
	"testing"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func TestParseCreateTaskText(t *testing.T) {
	command, err := parseCreateTaskText("write report 08100930 08101000", nil)
	if err != nil {
		t.Fatal(err)
	}

	if command.Title != "write report" {
		t.Fatalf("Title = %q, want %q", command.Title, "write report")
	}
	if command.Date == nil || command.Date.Start == nil || command.Date.Start.Format("01-02 15:04") != "08-10 09:30" {
		t.Fatalf("Date = %+v", command.Date)
	}
	if command.Notification == nil || command.Notification.Start == nil || command.Notification.Start.Format("01-02 15:04") != "08-10 10:00" {
		t.Fatalf("Notification = %+v", command.Notification)
	}
}

func TestParseCreateTaskTextAppliesPriority(t *testing.T) {
	command, err := parseCreateTaskText("call back 08100930 08101000", &taskdomain.SelectOption{Name: "High"})
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

func TestParseWorkText(t *testing.T) {
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	command, err := parseWorkText("start 08100930 08111845", now, "template-id")
	if err != nil {
		t.Fatal(err)
	}

	if command.Title != "ollo勤務" {
		t.Fatalf("Title = %q, want %q", command.Title, "ollo勤務")
	}
	if command.TemplateID != "template-id" {
		t.Fatalf("TemplateID = %q, want template-id", command.TemplateID)
	}
	if command.Status == nil || command.Status.Name != "inprogress" {
		t.Fatalf("Status = %+v", command.Status)
	}
	if command.Date == nil || command.Date.Start == nil || command.Date.End == nil {
		t.Fatalf("Date = %+v", command.Date)
	}
	if command.Date.Start.Format("2006-01-02 15:04") != "2026-08-10 09:30" {
		t.Fatalf("Date.Start = %s", command.Date.Start.Format("2006-01-02 15:04"))
	}
	if command.Date.End.Format("2006-01-02 15:04") != "2026-08-11 18:45" {
		t.Fatalf("Date.End = %s", command.Date.End.Format("2006-01-02 15:04"))
	}
}

func TestWorkStatusMapsInputs(t *testing.T) {
	tests := map[string]string{
		"start": "inprogress",
		"end":   "done",
		"todo":  "todo",
	}
	for input, want := range tests {
		got, err := workStatus(input)
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != want {
			t.Fatalf("workStatus(%q) = %q, want %q", input, got.Name, want)
		}
	}
}
