package notion

type TaskColumns struct {
	Title        string
	Status       string
	Date         string
	Category     string
	Label        string
	Priority     string
	Notification string
}

var OverviewTaskColumns = TaskColumns{
	Title:        "title",
	Status:       "status",
	Date:         "date",
	Category:     "category",
	Label:        "label",
	Priority:     "priority",
	Notification: "notification",
}
