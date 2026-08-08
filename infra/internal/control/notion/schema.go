package notion

type TaskColumns struct {
	DisplayID    string
	Title        string
	Status       string
	Date         string
	Category     string
	Label        string
	Priority     string
	Notification string
}

var MedControlTaskColumns = TaskColumns{
	DisplayID:    "No",
	Title:        "title",
	Status:       "status",
	Date:         "date",
	Category:     "category",
	Label:        "label",
	Priority:     "priority",
	Notification: "notification",
}
