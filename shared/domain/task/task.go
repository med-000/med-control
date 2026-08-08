package task

import "time"

type Task struct {
	ID           string         `json:"id"`
	DisplayID    string         `json:"display_id,omitempty"`
	Title        string         `json:"title"`
	Status       *SelectOption  `json:"status,omitempty"`
	Date         *DateRange     `json:"date,omitempty"`
	Categories   []SelectOption `json:"categories,omitempty"`
	Label        *SelectOption  `json:"label,omitempty"`
	Priority     *SelectOption  `json:"priority,omitempty"`
	Notification *DateRange     `json:"notification,omitempty"`
	Source       Source         `json:"source"`
	SourceID     string         `json:"source_id"`
	SourceURL    string         `json:"source_url,omitempty"`
	Description  string         `json:"description,omitempty"`
}

type SelectOption struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type DateRange struct {
	Start    *time.Time `json:"start,omitempty"`
	End      *time.Time `json:"end,omitempty"`
	TimeZone string     `json:"time_zone,omitempty"`
}

func (item Task) DisplayTitle() string {
	if item.DisplayID == "" {
		return item.Title
	}
	return "[" + item.DisplayID + "] " + item.Title
}

type CreateCommand struct {
	DisplayID    string         `json:"display_id,omitempty"`
	Title        string         `json:"title"`
	Status       *SelectOption  `json:"status,omitempty"`
	Date         *DateRange     `json:"date,omitempty"`
	Categories   []SelectOption `json:"categories,omitempty"`
	Label        *SelectOption  `json:"label,omitempty"`
	Priority     *SelectOption  `json:"priority,omitempty"`
	Notification *DateRange     `json:"notification,omitempty"`
	Description  string         `json:"description,omitempty"`
	TemplateID   string         `json:"template_id,omitempty"`
}

type UpdateCommand struct {
	DisplayID    *string         `json:"display_id,omitempty"`
	Title        *string         `json:"title,omitempty"`
	Status       *SelectOption   `json:"status,omitempty"`
	Date         *DateRange      `json:"date,omitempty"`
	Categories   *[]SelectOption `json:"categories,omitempty"`
	Label        *SelectOption   `json:"label,omitempty"`
	Priority     *SelectOption   `json:"priority,omitempty"`
	Notification *DateRange      `json:"notification,omitempty"`
	Description  *string         `json:"description,omitempty"`
}

type Filter struct {
	DisplayIDs  []string   `json:"display_ids,omitempty"`
	StatusIDs   []string   `json:"status_ids,omitempty"`
	StatusNames []string   `json:"status_names,omitempty"`
	LabelIDs    []string   `json:"label_ids,omitempty"`
	LabelNames  []string   `json:"label_names,omitempty"`
	PriorityIDs []string   `json:"priority_ids,omitempty"`
	CategoryIDs []string   `json:"category_ids,omitempty"`
	From        *time.Time `json:"from,omitempty"`
	To          *time.Time `json:"to,omitempty"`
}
