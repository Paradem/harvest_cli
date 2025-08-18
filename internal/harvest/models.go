package harvest

// Project represents a Harvest project.
type Project struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// Task represents a Harvest task.
type Task struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Task represents a Harvest task.
type TaskAssignment struct {
	Task Task `json:"task"`
}

// func (ta TaskAssignment) Name() string { return ta.Task.Name }
// func (ta TaskAssignment) ID() int64    { return ta.Task.ID }

// TimeEntryRequest is the payload for creating a time entry.
type TimeEntryRequest struct {
	ProjectID int64  `json:"project_id"`
	TaskID    int64  `json:"task_id"`
	SpendDate string `json:"spent_date"`
	Notes     string `json:"notes"`
}

// TimeEntryResponse represents the API response for a created time entry.
type TimeEntryResponse struct {
	ID        int64   `json:"id"`
	SpentDate string  `json:"spent_date"`
	Project   Project `json:"project"`
	Task      Task    `json:"task"`
}
