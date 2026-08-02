package notion

import (
	"fmt"
	"time"

	taskdomain "github.com/med-000/overview/shared/domain/task"
)

func RowsToTasks(rows []map[string]any) []taskdomain.Task {
	tasks := make([]taskdomain.Task, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, RowToTask(row))
	}
	return tasks
}

func RowToTask(row map[string]any) taskdomain.Task {
	id := stringValue(row["id"])
	properties := mapValue(row["properties"])

	return taskdomain.Task{
		ID:           fmt.Sprintf("%s:%s", taskdomain.SourceNotion, id),
		Title:        firstNonEmptyString(titleProperty(properties[OverviewTaskColumns.Title]), "Untitled"),
		Status:       selectLikeOption(properties[OverviewTaskColumns.Status]),
		Date:         dateProperty(properties[OverviewTaskColumns.Date]),
		Categories:   multiSelectOptions(properties[OverviewTaskColumns.Category]),
		Label:        selectLikeOption(properties[OverviewTaskColumns.Label]),
		Priority:     selectLikeOption(properties[OverviewTaskColumns.Priority]),
		Notification: dateProperty(properties[OverviewTaskColumns.Notification]),
		Source:       taskdomain.SourceNotion,
		SourceID:     id,
		SourceURL:    stringValue(row["url"]),
		Description:  stringValue(row["description"]),
	}
}

func titleProperty(value any) string {
	property := mapValue(value)
	if title := richTextPlainText(property["title"]); title != "" {
		return title
	}
	return ""
}

func selectLikeOption(value any) *taskdomain.SelectOption {
	property := mapValue(value)
	for _, key := range []string{"status", "select"} {
		option := selectOption(property[key])
		if option != nil {
			return option
		}
	}
	return nil
}

func selectLikeOptions(value any) []taskdomain.SelectOption {
	option := selectLikeOption(value)
	if option == nil {
		return nil
	}
	return []taskdomain.SelectOption{*option}
}

func multiSelectOptions(value any) []taskdomain.SelectOption {
	property := mapValue(value)
	options, ok := property["multi_select"].([]any)
	if !ok {
		return selectLikeOptions(value)
	}

	result := make([]taskdomain.SelectOption, 0, len(options))
	for _, optionValue := range options {
		option := selectOption(optionValue)
		if option != nil {
			result = append(result, *option)
		}
	}
	return result
}

func selectOption(value any) *taskdomain.SelectOption {
	option := mapValue(value)
	if option == nil {
		return nil
	}

	name := stringValue(option["name"])
	if name == "" {
		return nil
	}

	return &taskdomain.SelectOption{
		ID:    stringValue(option["id"]),
		Name:  name,
		Color: stringValue(option["color"]),
	}
}

func dateProperty(value any) *taskdomain.DateRange {
	property := mapValue(value)
	date := mapValue(property["date"])
	if date == nil {
		return nil
	}

	start := parseNotionTime(stringValue(date["start"]))
	end := parseNotionTime(stringValue(date["end"]))
	if start == nil && end == nil {
		return nil
	}

	return &taskdomain.DateRange{
		Start:    start,
		End:      end,
		TimeZone: stringValue(date["time_zone"]),
	}
}

func richTextPlainText(value any) string {
	texts, ok := value.([]any)
	if !ok {
		return ""
	}

	var result string
	for _, text := range texts {
		textMap := mapValue(text)
		result += stringValue(textMap["plain_text"])
	}
	return result
}

func parseNotionTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}

	return nil
}

func mapValue(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
