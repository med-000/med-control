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

func TestTemplatesFromResponse(t *testing.T) {
	templates, err := templatesFromResponse(map[string]any{
		"templates": []any{
			map[string]any{
				"id":         "template-id",
				"name":       "Work",
				"is_default": false,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(templates) != 1 {
		t.Fatalf("len(templates) = %d, want 1", len(templates))
	}
	if templates[0].ID != "template-id" || templates[0].Name != "Work" || templates[0].IsDefault {
		t.Fatalf("template = %+v", templates[0])
	}
}

func TestCreateDatabaseTaskPropertiesIncludeStatusAndTemplate(t *testing.T) {
	start := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)
	command := taskdomain.CreateCommand{
		Title:      "ollo勤務",
		Status:     &taskdomain.SelectOption{Name: "inprogress"},
		Date:       &taskdomain.DateRange{Start: &start, End: &end},
		TemplateID: "template-id",
	}

	properties := createTaskProperties(command)
	status := properties[MedControlTaskColumns.Status].(map[string]any)["status"].(map[string]any)
	if status["name"] != "inprogress" {
		t.Fatalf("status name = %q", status["name"])
	}

	requestBody := createTaskRequestBody("data-source-id", properties, command.TemplateID)
	template := requestBody["template"].(map[string]any)
	if template["template_id"] != "template-id" {
		t.Fatalf("template_id = %q", template["template_id"])
	}
}

func TestNotionStatusProperty(t *testing.T) {
	property := notionStatusProperty("done")
	status := property["status"].(map[string]any)

	if status["name"] != "done" {
		t.Fatalf("status name = %q", status["name"])
	}
}
