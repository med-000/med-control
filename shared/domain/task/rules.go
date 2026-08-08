package task

import (
	"strings"
	"time"
)

func MatchesFilter(item Task, filter Filter) bool {
	if len(filter.DisplayIDs) > 0 && !displayIDIn(item.DisplayID, filter.DisplayIDs) {
		return false
	}
	if len(filter.StatusIDs) > 0 && !selectIDIn(item.Status, filter.StatusIDs) {
		return false
	}
	if len(filter.StatusNames) > 0 && !selectNameIn(item.Status, filter.StatusNames) {
		return false
	}
	if len(filter.LabelIDs) > 0 && !selectIDIn(item.Label, filter.LabelIDs) {
		return false
	}
	if len(filter.LabelNames) > 0 && !selectNameIn(item.Label, filter.LabelNames) {
		return false
	}
	if len(filter.PriorityIDs) > 0 && !selectIDIn(item.Priority, filter.PriorityIDs) {
		return false
	}
	if len(filter.CategoryIDs) > 0 && !selectIDsInclude(item.Categories, filter.CategoryIDs) {
		return false
	}
	if filter.From != nil && (item.Date == nil || item.Date.Start == nil || item.Date.Start.Before(*filter.From)) {
		return false
	}
	if filter.To != nil && (item.Date == nil || item.Date.Start == nil || item.Date.Start.After(*filter.To)) {
		return false
	}
	return true
}

func DisplayIDMatches(displayID string, query string) bool {
	displayID = strings.ToLower(strings.TrimSpace(displayID))
	query = strings.ToLower(strings.TrimSpace(query))
	if displayID == "" || query == "" {
		return false
	}
	if displayID == query {
		return true
	}

	hyphen := strings.LastIndex(displayID, "-")
	return hyphen >= 0 && displayID[hyphen+1:] == query
}

func NotificationKey(item Task) string {
	if item.Notification == nil || item.Notification.Start == nil {
		return item.ID
	}
	return item.ID + ":" + item.Notification.Start.Format(time.RFC3339Nano)
}

func SameNotificationStart(left *DateRange, right *DateRange) bool {
	if left == nil || right == nil || left.Start == nil || right.Start == nil {
		return false
	}
	return left.Start.Equal(*right.Start)
}

func displayIDIn(displayID string, ids []string) bool {
	for _, id := range ids {
		if DisplayIDMatches(displayID, id) {
			return true
		}
	}
	return false
}

func selectIDIn(option *SelectOption, ids []string) bool {
	if option == nil {
		return false
	}
	for _, id := range ids {
		if option.ID == id {
			return true
		}
	}
	return false
}

func selectNameIn(option *SelectOption, names []string) bool {
	if option == nil {
		return false
	}
	for _, name := range names {
		if option.Name == name {
			return true
		}
	}
	return false
}

func selectIDsInclude(options []SelectOption, ids []string) bool {
	for _, option := range options {
		for _, id := range ids {
			if option.ID == id {
				return true
			}
		}
	}
	return false
}
