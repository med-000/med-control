package notion

import (
	"testing"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

func TestRowToTaskMapsUniqueIDWithoutPrefix(t *testing.T) {
	row := map[string]any{
		"id": "page-id",
		"properties": map[string]any{
			OverviewTaskColumns.DisplayID: map[string]any{
				"unique_id": map[string]any{
					"number": float64(5),
				},
			},
			OverviewTaskColumns.Title: map[string]any{
				"title": []any{
					map[string]any{"plain_text": "ollo写真撮影"},
				},
			},
		},
	}

	got := RowToTask(row)

	if got.DisplayID != "5" {
		t.Fatalf("DisplayID = %q, want %q", got.DisplayID, "5")
	}
	if got.ID != string(taskdomain.SourceNotion)+":page-id" {
		t.Fatalf("ID = %q", got.ID)
	}
}

func TestRowToTaskMapsUniqueIDWithPrefix(t *testing.T) {
	row := map[string]any{
		"id": "page-id",
		"properties": map[string]any{
			OverviewTaskColumns.DisplayID: map[string]any{
				"unique_id": map[string]any{
					"prefix": "OVW",
					"number": float64(6),
				},
			},
		},
	}

	got := RowToTask(row)

	if got.DisplayID != "OVW-6" {
		t.Fatalf("DisplayID = %q, want %q", got.DisplayID, "OVW-6")
	}
}
