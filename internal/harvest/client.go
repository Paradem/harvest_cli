package harvest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://api.harvestapp.com/v2"

// Client holds the HTTP client and auth info.
type Client struct {
	httpClient *http.Client
	accountID  string
	token      string
}

// NewClient creates a Harvest API client using the provided account ID and access token.
func NewClient(accountID, accessToken string) (*Client, error) {
	if accountID == "" || accessToken == "" {
		return nil, fmt.Errorf("account ID and access token must be provided")
	}
	return &Client{httpClient: http.DefaultClient, accountID: accountID, token: accessToken}, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewReader(b)
	}
	url := fmt.Sprintf("%s%s", baseURL, path)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Harvest-Account-ID", c.accountID)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("harvest API error %d: %s", resp.StatusCode, string(body))
	}
	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}

// ListProjects fetches all projects.
func (c *Client) ListProjects() ([]Project, error) {
	req, err := c.newRequest("GET", "/projects?is_active=true", nil)
	if err != nil {
		return nil, err
	}
	var res struct {
		Projects []Project `json:"projects"`
	}
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return res.Projects, nil
}

func pluckTasks(ts []TaskAssignment) []Task {
	tasks := make([]Task, 0, len(ts)) // pre‑allocate capacity
	for _, v := range ts {
		tasks = append(tasks, v.Task)
	}
	return tasks
}

// ListTasks fetches tasks for a project.
func (c *Client) ListTasks(projectID int64) ([]Task, error) {
	path := fmt.Sprintf("/projects/%d/task_assignments", projectID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		TaskAssignments []TaskAssignment `json:"task_assignments"`
	}

	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return pluckTasks(res.TaskAssignments), nil
}

// CreateTimeEntry posts a new time entry.
func (c *Client) CreateTimeEntry(entry TimeEntryRequest) (*TimeEntryResponse, error) {
	req, err := c.newRequest("POST", "/time_entries", entry)
	if err != nil {
		return nil, err
	}
	var res TimeEntryResponse

	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ListTimeEntries fetches time entries with optional date and user filtering.
func (c *Client) ListTimeEntries(from, to *string, userID *int64) ([]TimeEntry, error) {
	path := "/time_entries"
	params := make([]string, 0, 3)
	if from != nil {
		params = append(params, "from="+*from)
	}
	if to != nil {
		params = append(params, "to="+*to)
	}
	if userID != nil {
		params = append(params, fmt.Sprintf("user_id=%d", *userID))
	}

	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		TimeEntries []TimeEntry `json:"time_entries"`
	}
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return res.TimeEntries, nil
}

// RestartTimeEntry restarts a stopped time entry.
func (c *Client) RestartTimeEntry(timeEntryID int64) (*TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d/restart", timeEntryID)
	req, err := c.newRequest("PATCH", path, nil)
	if err != nil {
		return nil, err
	}

	var res TimeEntry
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
