package tasksync

import (
	"testing"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func TestShouldCompleteScheduleItem(t *testing.T) {
	start := time.Date(2026, 8, 8, 9, 30, 0, 0, time.Local)
	task := taskdomain.Task{
		Label: &taskdomain.SelectOption{Name: "schedule"},
		Date:  &taskdomain.DateRange{Start: &start},
	}
	now := time.Date(2026, 8, 10, 1, 0, 0, 0, time.Local)

	if !shouldCompleteScheduleItem(task, now) {
		t.Fatal("schedule item should be completed")
	}
}

func TestShouldCompleteScheduleItemRejectsOtherLabel(t *testing.T) {
	start := time.Date(2026, 8, 8, 9, 30, 0, 0, time.Local)
	task := taskdomain.Task{
		Label: &taskdomain.SelectOption{Name: "work"},
		Date:  &taskdomain.DateRange{Start: &start},
	}
	now := time.Date(2026, 8, 10, 1, 0, 0, 0, time.Local)

	if shouldCompleteScheduleItem(task, now) {
		t.Fatal("non-schedule item should not be completed")
	}
}

func TestShouldCompleteScheduleItemRejectsDone(t *testing.T) {
	start := time.Date(2026, 8, 8, 9, 30, 0, 0, time.Local)
	task := taskdomain.Task{
		Label:  &taskdomain.SelectOption{Name: "schedule"},
		Status: &taskdomain.SelectOption{Name: "done"},
		Date:   &taskdomain.DateRange{Start: &start},
	}
	now := time.Date(2026, 8, 10, 1, 0, 0, 0, time.Local)

	if shouldCompleteScheduleItem(task, now) {
		t.Fatal("done schedule item should not be completed")
	}
}

func TestShouldCompleteScheduleItemRejectsToday(t *testing.T) {
	start := time.Date(2026, 8, 10, 9, 30, 0, 0, time.Local)
	task := taskdomain.Task{
		Label: &taskdomain.SelectOption{Name: "schedule"},
		Date:  &taskdomain.DateRange{Start: &start},
	}
	now := time.Date(2026, 8, 10, 1, 0, 0, 0, time.Local)

	if shouldCompleteScheduleItem(task, now) {
		t.Fatal("today's schedule item should not be completed")
	}
}
