package harvest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const baseURL = "https://api.harvestapp.com/v2"

// Client holds the HTTP client and auth info.
type Client struct {
	httpClient *http.Client
	accountID  string
	token      string
}

// NewClient creates a Harvest API client using env vars HARVEST_ACCOUNT_ID and HARVEST_ACCESS_TOKEN.
func NewClient() (*Client, error) {
	acc := os.Getenv("HARVEST_ACCOUNT_ID")
	tok := os.Getenv("HARVEST_ACCESS_TOKEN")
	if acc == "" || tok == "" {
		return nil, fmt.Errorf("environment variables HARVEST_ACCOUNT_ID and HARVEST_ACCESS_TOKEN must be set")
	}
	return &Client{httpClient: http.DefaultClient, accountID: acc, token: tok}, nil
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
