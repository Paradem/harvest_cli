package harvest

// Project represents a Harvest project.
type Project struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name"`
	Active bool          `json:"active"`
	Client HarvestClient `json:"client"`
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

// TimeEntry represents a Harvest time entry from the API.
type TimeEntry struct {
	ID                int64              `json:"id"`
	SpentDate         string             `json:"spent_date"`
	User              User               `json:"user"`
	Client            HarvestClient      `json:"client"`
	Project           Project            `json:"project"`
	Task              Task               `json:"task"`
	UserAssignment    UserAssignment     `json:"user_assignment"`
	TaskAssignment    TaskAssignment     `json:"task_assignment"`
	Hours             float64            `json:"hours"`
	HoursWithoutTimer float64            `json:"hours_without_timer"`
	RoundedHours      float64            `json:"rounded_hours"`
	Notes             *string            `json:"notes"`
	IsLocked          bool               `json:"is_locked"`
	LockedReason      *string            `json:"locked_reason"`
	IsClosed          bool               `json:"is_closed"`
	ApprovalStatus    string             `json:"approval_status"`
	IsBilled          bool               `json:"is_billed"`
	TimerStartedAt    *string            `json:"timer_started_at"`
	StartedTime       *string            `json:"started_time"`
	EndedTime         *string            `json:"ended_time"`
	IsRunning         bool               `json:"is_running"`
	Invoice           *Invoice           `json:"invoice"`
	ExternalReference *ExternalReference `json:"external_reference"`
	Billable          bool               `json:"billable"`
	Budgeted          bool               `json:"budgeted"`
	BillableRate      *float64           `json:"billable_rate"`
	CostRate          *float64           `json:"cost_rate"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
}

// User represents a Harvest user.
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// HarvestClient represents a Harvest client.
type HarvestClient struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// UserAssignment represents a user assignment to a project.
type UserAssignment struct {
	ID               int64    `json:"id"`
	IsProjectManager bool     `json:"is_project_manager"`
	IsActive         bool     `json:"is_active"`
	Budget           *float64 `json:"budget"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	HourlyRate       *float64 `json:"hourly_rate"`
}

// Invoice represents an invoice.
type Invoice struct {
	ID     int64  `json:"id"`
	Number string `json:"number"`
}

// ExternalReference represents an external reference.
type ExternalReference struct {
	ID             string `json:"id"`
	GroupID        string `json:"group_id"`
	AccountID      string `json:"account_id"`
	Permalink      string `json:"permalink"`
	Service        string `json:"service"`
	ServiceIconURL string `json:"service_icon_url"`
}
