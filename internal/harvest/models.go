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
	ID                int64          `json:"id"`
	SpentDate         string         `json:"spent_date"`
	User              User           `json:"user"`
	Client            ClientInfo     `json:"client"`
	Project           ProjectDetail  `json:"project"`
	Task              TaskDetail     `json:"task"`
	UserAssign        UserAssignment `json:"user_assignment"`
	TaskAssign        TaskAssignment `json:"task_assignment"`
	Hours             float64        `json:"hours"`
	Rounded           float64        `json:"rounded_hours"`
	Notes             *string        `json:"notes"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
	IsLocked          bool           `json:"is_locked"`
	LockedReason      *string        `json:"locked_reason"`
	IsClosed          bool           `json:"is_closed"`
	ApprovalStatus    string         `json:"approval_status"`
	IsBilled          bool           `json:"is_billed"`
	TimerStartedAt    *string        `json:"timer_started_at"`
	StartedTime       *string        `json:"started_time"`
	EndedTime         *string        `json:"ended_time"`
	IsRunning         bool           `json:"is_running"`
	Invoice           interface{}    `json:"invoice"`
	ExternalReference *string        `json:"external_reference"`
	Billable          bool           `json:"billable"`
	Budgeted          bool           `json:"budgeted"`
	BillableRate      float64        `json:"billable_rate"`
	CostRate          float64        `json:"cost_rate"`
}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type ClientInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type ProjectDetail struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type TaskDetail struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type UserAssignment struct {
	ID               int64       `json:"id"`
	IsProjectManager bool        `json:"is_project_manager"`
	IsActive         bool        `json:"is_active"`
	Budget           interface{} `json:"budget"`
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
	HourlyRate       float64     `json:"hourly_rate"`
}

type TaskAssignment struct {
	ID         int64       `json:"id"`
	Billable   bool        `json:"billable"`
	IsActive   bool        `json:"is_active"`
	CreatedAt  string      `json:"created_at"`
	UpdatedAt  string      `json:"updated_at"`
	HourlyRate float64     `json:"hourly_rate"`
	Budget     interface{} `json:"budget"`
}
