package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/harvestcli/internal/config"
	"github.com/example/harvestcli/internal/harvest"
	"github.com/example/harvestcli/internal/prompt"
)

func setupGlobalConfig(cfg *config.Config) error {
	fmt.Println("Harvest CLI needs to be configured. Please provide the following information:")
	fmt.Println()

	// Prompt for account ID
	accountID, err := prompt.InputPrompt("Harvest Account ID:", "")
	if err != nil {
		return fmt.Errorf("failed to get account ID: %v", err)
	}
	if accountID == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	// Prompt for access token
	accessToken, err := prompt.InputPrompt("Harvest Access Token:", "")
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	if accessToken == "" {
		return fmt.Errorf("access token cannot be empty")
	}

	// Prompt for user ID
	userID, err := prompt.InputPrompt("Harvest User ID:", "")
	if err != nil {
		return fmt.Errorf("failed to get user ID: %v", err)
	}
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Update config
	cfg.HarvestAccountID = accountID
	cfg.HarvestAccessToken = accessToken
	cfg.HarvestUserID = userID

	// Save config
	if err := cfg.SaveGlobal(); err != nil {
		return fmt.Errorf("failed to save global config: %v", err)
	}

	fmt.Println("Configuration saved successfully!")
	return nil
}

func handleStopTimer(client *harvest.Client, userIDStr string, logger *log.Logger) {
	if userIDStr == "" {
		logger.Fatalf("User ID must be provided")
		os.Exit(1)
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Fatalf("Invalid user ID: %v", err)
		os.Exit(1)
	}

	// Get today's date for filtering
	today := time.Now().Format("2006-01-02")

	// List time entries for today filtered by current user
	entries, err := client.ListTimeEntries(&today, &today, &userID)
	if err != nil {
		logger.Fatalf("Failed to list time entries: %v", err)
		os.Exit(1)
	}

	// Find running entries
	var runningEntry *harvest.TimeEntry
	for _, entry := range entries {
		if entry.IsRunning {
			runningEntry = &entry
			break // Take the first running entry (there should typically be only one)
		}
	}

	if runningEntry == nil {
		fmt.Println("No running timer found to stop.")
		return
	}

	// Stop the time entry
	stoppedEntry, err := client.StopTimeEntry(runningEntry.ID)
	if err != nil {
		logger.Fatalf("Failed to stop time entry: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Stopped time entry %d for project %s task %s\n",
		stoppedEntry.ID, stoppedEntry.Project.Name, stoppedEntry.Task.Name)
}

func handleTimeEntrySelection(client *harvest.Client, userIDStr string, logger *log.Logger) {
	if userIDStr == "" {
		logger.Fatalf("User ID must be provided")
		os.Exit(1)
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Fatalf("Invalid user ID: %v", err)
		os.Exit(1)
	}

	// Get today's date for filtering
	today := time.Now().Format("2006-01-02")

	// List time entries for today filtered by current user
	entries, err := client.ListTimeEntries(&today, &today, &userID)
	if err != nil {
		logger.Fatalf("Failed to list time entries: %v", err)
		os.Exit(1)
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

		// Convert decimal hours to [HH:MM] format
		totalHours := entry.Hours
		hours := int(totalHours)
		minutes := int(math.Ceil((totalHours - float64(hours)) * 60))
		// Handle case where minutes rounds up to 60 (should increment hours)
		if minutes >= 60 {
			hours++
			minutes = 0
		}

		entryOptions[i] = fmt.Sprintf("%s - %s (%s) [%02d:%02d]%s",
			entry.Project.Name, entry.Task.Name, status, hours, minutes, notes)
	}

	// Show selection prompt
	idx, err := prompt.SelectPrompt(entryOptions, "Select a time entry to restart:")
	if err != nil {
		logger.Fatalf("Selection error: %v", err)
		os.Exit(1)
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
		logger.Fatalf("Failed to restart time entry: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Restarted time entry %d for project %s task %s\n",
		restartedEntry.ID, restartedEntry.Project.Name, restartedEntry.Task.Name)
}

func handleStatusDisplay(client *harvest.Client, userIDStr string, logger *log.Logger, sketchyBarMode bool) {
	if userIDStr == "" {
		logger.Fatalf("User ID must be provided")
		os.Exit(1)
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Fatalf("Invalid user ID: %v", err)
		os.Exit(1)
	}

	// Get today's date for filtering
	today := time.Now().Format("2006-01-02")

	// List time entries for today filtered by current user
	entries, err := client.ListTimeEntries(&today, &today, &userID)
	if err != nil {
		logger.Fatalf("Failed to list time entries: %v", err)
		os.Exit(1)
	}

	// Find running entries
	var runningEntry *harvest.TimeEntry
	for _, entry := range entries {
		if entry.IsRunning {
			runningEntry = &entry
			break // Take the first running entry (there should typically be only one)
		}
	}

	if runningEntry == nil {
		// No running timer - show total billable hours for today
		var totalBillableHours float64
		for _, entry := range entries {
			if entry.Billable {
				totalBillableHours += entry.Hours
			}
		}

		// Convert decimal hours to [HH:MM] format
		hours := int(totalBillableHours)
		minutes := int(math.Ceil((totalBillableHours - float64(hours)) * 60))

		// Handle case where minutes rounds up to 60 (should increment hours)
		if minutes >= 60 {
			hours++
			minutes = 0
		}

		// Display total billable hours with [HH:MM] format in green (same as running timer)
		if sketchyBarMode {
			fmt.Printf("[%02d:%02d] paused\n", hours, minutes)
		} else {
			fmt.Printf("#[fg=colour46][%02d:%02d]#[default] paused", hours, minutes)
		}
		return
	}

	// Use total hours from the time entry (includes all accumulated time)
	totalHours := runningEntry.Hours
	hours := int(totalHours)
	minutes := int(math.Ceil((totalHours - float64(hours)) * 60))

	// Handle case where minutes rounds up to 60 (should increment hours)
	if minutes >= 60 {
		hours++
		minutes = 0
	}

	// Prepare notes display (first word of first line only)
	notesDisplay := ""
	if runningEntry.Notes != nil && *runningEntry.Notes != "" {
		// Split by newline and take first line
		lines := strings.Split(*runningEntry.Notes, "\n")
		firstLine := strings.TrimSpace(lines[0])

		// Split by space and take first word
		words := strings.Fields(firstLine)
		if len(words) > 0 {
			notesDisplay = " " + words[0]
		}
	}

	// Display running timer with [HH:MM] format in green
	if sketchyBarMode {
		fmt.Printf("[%02d:%02d]%s\n", hours, minutes, notesDisplay)
	} else {
		fmt.Printf("#[fg=colour46][%02d:%02d]#[default]%s",
			hours, minutes, notesDisplay)
	}
}

func main() {
	// Setup logger
	logger, err := config.SetupLogger()
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}

	var note string
	var configPath string
	var ignoreConfig bool
	var selectEntry bool
	var showStatus bool
	var sketchyBarMode bool
	var stopTimer bool
	flag.StringVar(&note, "n", "", "Initial notes text")
	flag.StringVar(&configPath, "c", config.DefaultConfigPath(), "Config file path")
	flag.BoolVar(&ignoreConfig, "i", false, "Ignore loading local configuration")
	flag.BoolVar(&selectEntry, "e", false, "Select and restart an existing time entry")
	flag.BoolVar(&showStatus, "s", false, "Show current running timer status")
	flag.BoolVar(&sketchyBarMode, "b", false, "Format output for SketchyBar (plain text, must be used with -s)")
	flag.BoolVar(&stopTimer, "q", false, "Stop the currently running timer")
	var ticket string
	flag.StringVar(&ticket, "t", "", "External ticket number to prefix notes")
	flag.Parse()

	// Validate flags
	if sketchyBarMode && !showStatus {
		logger.Fatalf("-b flag must be used with -s flag")
		os.Exit(1)
	}

	// Load global configuration
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		logger.Fatalf("Failed to load global config: %v", err)
		os.Exit(1)
	}

	// Check if global config is complete, if not, prompt for setup
	if globalCfg.HarvestAccountID == "" || globalCfg.HarvestAccessToken == "" || globalCfg.HarvestUserID == "" {
		setupErr := setupGlobalConfig(globalCfg)
		if setupErr != nil {
			logger.Fatalf("Failed to setup global config: %v", setupErr)
			os.Exit(1)
		}
	}

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
	var loadErr error
	if !ignoreConfig {
		cfg, loadErr = config.Load(configPath)
	} else {
		cfg = &config.Config{}
	}

	if loadErr != nil {
		logger.Fatalf("Failed to load config: %v", loadErr)
		os.Exit(1)
	}

	client, clientErr := harvest.NewClient(globalCfg.HarvestAccountID, globalCfg.HarvestAccessToken)
	if clientErr != nil {
		logger.Fatalf("Auth error: %v", clientErr)
		os.Exit(1)
	}

	// Handle stop timer mode
	if stopTimer {
		handleStopTimer(client, globalCfg.HarvestUserID, logger)
		return
	}

	// Handle time entry selection mode
	if selectEntry {
		handleTimeEntrySelection(client, globalCfg.HarvestUserID, logger)
		return
	}

	// Handle status display mode
	if showStatus {
		handleStatusDisplay(client, globalCfg.HarvestUserID, logger, sketchyBarMode)
		return
	}

	// Projects selection
	projects, err := client.ListProjects()
	if err != nil {
		logger.Fatalf("Failed to list projects: %v", err)
		os.Exit(1)
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
			logger.Fatalf("prompt error: %v", err)
			os.Exit(1)
		}
		selectedProjectID = projects[idx].ID
	}

	// Tasks selection
	tasks, err := client.ListTasks(selectedProjectID)
	if err != nil {
		logger.Fatalf("Failed to list tasks: %v", err)
		os.Exit(1)
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
			logger.Fatalf("prompt error: %v", err)
			os.Exit(1)
		}
		selectedTaskID = tasks[idx].ID
	}

	// Notes input
	notes := note
	if notes == "" {
		var err error
		notes, err = prompt.InputPrompt("Enter notes:", "")
		if err != nil {
			logger.Fatalf("prompt error: %v", err)
			os.Exit(1)
		}
	}

	// Create time entry
	req := harvest.TimeEntryRequest{ProjectID: selectedProjectID, TaskID: selectedTaskID, SpendDate: time.Now().Format(time.RFC3339), Notes: notes}
	resp, err := client.CreateTimeEntry(req)
	if err != nil {
		logger.Fatalf("Failed to create time entry: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Created time entry ID %d for project %s task %s\n", resp.ID, resp.Project.Name, resp.Task.Name)

	// Save defaults
	cfg.ProjectID = selectedProjectID
	cfg.TaskID = selectedTaskID
	if !ignoreConfig {
		if err := cfg.Save(configPath); err != nil {
			logger.Printf("Failed to save config: %v", err)
		}
	}

}
