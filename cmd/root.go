package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/harvestcli/internal/config"
	"github.com/example/harvestcli/internal/harvest"
	"github.com/example/harvestcli/internal/prompt"
)

func handleTimeEntrySelection(client *harvest.Client) {
	// Get user ID from environment variable
	userIDStr := os.Getenv("HARVEST_USER_ID")
	if userIDStr == "" {
		log.Fatalf("HARVEST_USER_ID environment variable must be set")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid HARVEST_USER_ID: %v", err)
	}

	// Get today's date for filtering
	today := time.Now().Format("2006-01-02")

	// List time entries for today filtered by current user
	entries, err := client.ListTimeEntries(&today, &today, &userID)
	if err != nil {
		log.Fatalf("Failed to list time entries: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found for today.")
		return
	}

	// Create options for selection
	entryOptions := make([]string, len(entries))
	for i, entry := range entries {
		status := "\033[33mStopped\033[0m" // Yellow for stopped
		if entry.IsRunning {
			status = "\033[32mRunning\033[0m" // Green for running
		}
		notes := ""
		if entry.Notes != nil {
			// Replace newlines with spaces and clean up formatting
			notes = strings.ReplaceAll(*entry.Notes, "\n", " | ")
			notes = strings.ReplaceAll(notes, "\r", " | ")
			// Trim whitespace and add padding
			notes = strings.TrimSpace(notes)
			// Truncate very long notes to prevent wrapping issues
			if len(notes) > 60 {
				notes = notes[:57] + "..."
			}
			if notes != "" {
				// Add cyan color highlighting for notes
				notes = fmt.Sprintf("  \033[36m%s\033[0m", notes)
			}
		}
		entryOptions[i] = fmt.Sprintf("%s - %s (%s) [%.2fh]%s",
			entry.Project.Name, entry.Task.Name, status, entry.Hours, notes)
	}

	// Show selection prompt
	idx, err := prompt.SelectPrompt(entryOptions, "Select a time entry to restart:")
	if err != nil {
		log.Fatalf("Selection error: %v", err)
	}

	selectedEntry := entries[idx]

	// Check if entry is already running
	if selectedEntry.IsRunning {
		fmt.Printf("Time entry %d is already running.\n", selectedEntry.ID)
		return
	}

	// Restart the time entry
	restartedEntry, err := client.RestartTimeEntry(selectedEntry.ID)
	if err != nil {
		log.Fatalf("Failed to restart time entry: %v", err)
	}

	fmt.Printf("Restarted time entry %d for project %s task %s\n",
		restartedEntry.ID, restartedEntry.Project.Name, restartedEntry.Task.Name)
}

func main() {
	var note string
	var configPath string
	var ignoreConfig bool
	var selectEntry bool
	flag.StringVar(&note, "n", "", "Initial notes text")
	flag.StringVar(&configPath, "c", config.DefaultConfigPath(), "Config file path")
	flag.BoolVar(&ignoreConfig, "i", false, "Ignore loading local configuration")
	flag.BoolVar(&selectEntry, "e", false, "Select and restart an existing time entry")
	var ticket string
	flag.StringVar(&ticket, "t", "", "External ticket number to prefix notes")
	flag.Parse()

	// Combine ticket with note if provided
	if ticket != "" {
		// Ensure ticket starts with '#'
		if !strings.HasPrefix(ticket, "#") {
			ticket = "#" + ticket
		}
		prefix := fmt.Sprintf("%s\n", ticket)
		if note != "" {
			note = prefix + note
		} else {
			note = prefix
		}
	}

	var cfg *config.Config
	var err error
	if !ignoreConfig {
		cfg, err = config.Load(configPath)
	} else {
		cfg = &config.Config{}
	}

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := harvest.NewClient()
	if err != nil {
		log.Fatalf("Auth error: %v", err)
	}

	// Handle time entry selection mode
	if selectEntry {
		handleTimeEntrySelection(client)
		return
	}

	// Projects selection
	projects, err := client.ListProjects()
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}
	projectOptions := make([]string, len(projects))
	for i, p := range projects {
		projectOptions[i] = fmt.Sprintf("%s \033[36m(%s)\033[0m", p.Name, p.Client.Name)
	}

	var selectedProjectID int64
	if cfg.ProjectID != 0 {
		// verify exists in list
		found := false
		for _, p := range projects {
			if p.ID == cfg.ProjectID {
				found = true
				break
			}
		}
		if found {
			selectedProjectID = cfg.ProjectID
		}
	}
	if selectedProjectID == 0 {
		var err error
		idx, err := prompt.SelectPrompt(projectOptions, "Select a project:")
		if err != nil {
			log.Fatalf("prompt error: %v", err)
		}
		selectedProjectID = projects[idx].ID
	}

	// Tasks selection
	tasks, err := client.ListTasks(selectedProjectID)
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	taskOptions := make([]string, len(tasks))
	for i, t := range tasks {
		taskOptions[i] = t.Name
	}

	var selectedTaskID int64
	if cfg.TaskID != 0 {
		found := false
		for _, t := range tasks {
			if t.ID == cfg.TaskID {
				found = true
				break
			}
		}
		if found {
			selectedTaskID = cfg.TaskID
		}
	}
	if selectedTaskID == 0 {
		var err error
		idx, err := prompt.SelectPrompt(taskOptions, "Select a task:")
		if err != nil {
			log.Fatalf("prompt error: %v", err)
		}
		selectedTaskID = tasks[idx].ID
	}

	// Notes input
	notes := note
	if notes == "" {
		var err error
		notes, err = prompt.InputPrompt("Enter notes:", "")
		if err != nil {
			log.Fatalf("prompt error: %v", err)
		}
	}

	// Create time entry
	req := harvest.TimeEntryRequest{ProjectID: selectedProjectID, TaskID: selectedTaskID, SpendDate: time.Now().Format(time.RFC3339), Notes: notes}
	resp, err := client.CreateTimeEntry(req)
	if err != nil {
		log.Fatalf("Failed to create time entry: %v", err)
	}

	fmt.Printf("Created time entry ID %d for project %s task %s\n", resp.ID, resp.Project.Name, resp.Task.Name)

	// Save defaults
	cfg.ProjectID = selectedProjectID
	cfg.TaskID = selectedTaskID
	if !ignoreConfig {
		if err := cfg.Save(configPath); err != nil {
			log.Printf("Failed to save config: %v", err)
		}
	}

}
