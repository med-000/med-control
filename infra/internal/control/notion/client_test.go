package notion

import (
	"testing"
	"time"

	taskdomain "github.com/med-000/med-control/shared/domain/task"
)

func TestNotionDateProperty(t *testing.T) {
	start := time.Date(2026, 8, 10, 9, 30, 0, 0, time.UTC)

	property := notionDateProperty(&taskdomain.DateRange{Start: &start})
	date, ok := property["date"].(map[string]any)
	if !ok {
		t.Fatalf("date property = %+v", property["date"])
	}

	if date["start"] != "2026-08-10T09:30:00Z" {
		t.Fatalf("start = %q", date["start"])
	}
}
