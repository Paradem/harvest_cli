package harvest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// StopTimeEntry stops a running time entry.
func (c *Client) StopTimeEntry(timeEntryID int64) (*TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d/stop", timeEntryID)
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

// UpdateTimeEntry updates a time entry with new hours.
func (c *Client) UpdateTimeEntry(timeEntryID int64, hours float64) (*TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d", timeEntryID)
	updateReq := TimeEntryUpdateRequest{Hours: hours}
	req, err := c.newRequest("PATCH", path, updateReq)
	if err != nil {
		return nil, err
	}

	var res TimeEntry
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) ListInvoices(from, to *string) ([]InvoiceDetail, error) {
	path := "/invoices?per_page=100"
	if from != nil {
		path += "&from=" + *from
	}
	if to != nil {
		path += "&to=" + *to
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var res InvoiceListResponse
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return res.Invoices, nil
}

func (c *Client) ListExpenses(from, to *string) ([]ExpenseDetail, error) {
	path := "/expenses"
	params := make([]string, 0, 2)
	if from != nil {
		params = append(params, "from="+*from)
	}
	if to != nil {
		params = append(params, "to="+*to)
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var res ExpenseListResponse
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return res.Expenses, nil
}

func (c *Client) ListExpenseCategories() ([]ExpenseCategory, error) {
	req, err := c.newRequest("GET", "/expense_categories?is_active=true", nil)
	if err != nil {
		return nil, err
	}

	var res ExpenseCategoryListResponse
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return res.ExpenseCategories, nil
}

func (c *Client) CreateExpense(reqBody ExpenseCreateRequest) (*ExpenseDetail, error) {
	req, err := c.newRequest("POST", "/expenses", reqBody)
	if err != nil {
		return nil, err
	}

	var res ExpenseDetail
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) CreateExpenseWithReceipt(reqBody ExpenseCreateRequest, receiptPath string) (*ExpenseDetail, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("project_id", strconv.FormatInt(reqBody.ProjectID, 10))
	writer.WriteField("expense_category_id", strconv.FormatInt(reqBody.ExpenseCategoryID, 10))
	writer.WriteField("spent_date", reqBody.SpentDate)
	writer.WriteField("total_cost", strconv.FormatFloat(reqBody.TotalCost, 'f', 2, 64))
	if reqBody.Notes != nil {
		writer.WriteField("notes", *reqBody.Notes)
	}
	if reqBody.Billable != nil {
		if *reqBody.Billable {
			writer.WriteField("billable", "true")
		} else {
			writer.WriteField("billable", "false")
		}
	}

	file, err := os.Open(receiptPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	part, err := writer.CreateFormFile("receipt", filepath.Base(receiptPath))
	if err != nil {
		return nil, err
	}
	io.Copy(part, file)
	writer.Close()

	url := baseURL + "/expenses"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Harvest-Account-ID", c.accountID)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var res ExpenseDetail
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
