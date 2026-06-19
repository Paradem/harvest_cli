package main

import (
	"encoding/json"
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

func handleStatusDisplay(client *harvest.Client, userIDStr string, logger *log.Logger, sketchyBarMode bool, waybarMode bool) {
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

		// Display total billable hours with [HH:MM] format
		if sketchyBarMode {
			fmt.Printf("[%02d:%02d] paused\n", hours, minutes)
		} else if waybarMode {
			fmt.Printf("{\"text\":\"<span color='#ff0000'>[%02d:%02d]</span> paused\",\"class\":\"paused\"}\n", hours, minutes)
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

	// Display running timer with [HH:MM] format
	if sketchyBarMode {
		fmt.Printf("[%02d:%02d]%s\n", hours, minutes, notesDisplay)
	} else if waybarMode {
		notesText := ""
		if notesDisplay != "" {
			notesText = fmt.Sprintf(" <span color='#ffffff'>%s</span>", notesDisplay[1:]) // Remove leading space
		}
		fmt.Printf("{\"text\":\"<span color='#00ff00'>[%02d:%02d]</span>%s\",\"class\":\"running\"}\n", hours, minutes, notesText)
	} else {
		fmt.Printf("#[fg=colour46][%02d:%02d]#[default]%s",
			hours, minutes, notesDisplay)
	}
}

func handleAddTime(client *harvest.Client, userIDStr string, logger *log.Logger, minutesToAdd int) {
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
		fmt.Println("No running timer found to add time to.")
		return
	}

	// Calculate new total hours
	additionalHours := float64(minutesToAdd) / 60.0
	newTotalHours := runningEntry.Hours + additionalHours

	// Update the time entry with new hours
	updatedEntry, err := client.UpdateTimeEntry(runningEntry.ID, newTotalHours)
	if err != nil {
		logger.Fatalf("Failed to update time entry: %v", err)
		os.Exit(1)
	}

	// Convert new total to [HH:MM] format for display
	totalHours := updatedEntry.Hours
	hours := int(totalHours)
	minutes := int(math.Ceil((totalHours - float64(hours)) * 60))

	// Handle case where minutes rounds up to 60 (should increment hours)
	if minutes >= 60 {
		hours++
		minutes = 0
	}

	fmt.Printf("Added %d minutes to running timer. New total: [%02d:%02d] for project %s task %s\n",
		minutesToAdd, hours, minutes, updatedEntry.Project.Name, updatedEntry.Task.Name)
}

func handleInvoiceList(client *harvest.Client, logger *log.Logger, from, to *string, jsonOutput bool) {
	invoices, err := client.ListInvoices(from, to)
	if err != nil {
		logger.Fatalf("Failed to list invoices: %v", err)
		os.Exit(1)
	}

	if jsonOutput {
		out, err := jsonMarshal(invoices)
		if err != nil {
			logger.Fatalf("Failed to marshal invoices: %v", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		return
	}

	if len(invoices) == 0 {
		fmt.Println("No invoices found.")
		return
	}

	fmt.Printf("%-12s %-12s %12s %-10s %s\n", "DATE", "NUMBER", "AMOUNT", "STATUS", "CLIENT")
	fmt.Println(strings.Repeat("-", 80))
	for _, inv := range invoices {
		issuedDate := inv.IssuedAt
		if issuedDate == "" {
			issuedDate = inv.CreatedAt
		}
		if len(issuedDate) > 10 {
			issuedDate = issuedDate[:10]
		}
		amount := fmt.Sprintf("$%.2f", inv.Amount)
		clientName := inv.Client.Name
		if len(clientName) > 30 {
			clientName = clientName[:27] + "..."
		}
		fmt.Printf("%-12s %-12s %12s %-10s %s\n",
			issuedDate, inv.Number, amount, inv.Status, clientName)
	}
}

func handleExpenseList(client *harvest.Client, logger *log.Logger, from, to *string, jsonOutput bool) {
	expenses, err := client.ListExpenses(from, to)
	if err != nil {
		logger.Fatalf("Failed to list expenses: %v", err)
		os.Exit(1)
	}

	if jsonOutput {
		out, err := jsonMarshal(expenses)
		if err != nil {
			logger.Fatalf("Failed to marshal expenses: %v", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		return
	}

	if len(expenses) == 0 {
		fmt.Println("No expenses found.")
		return
	}

	fmt.Printf("%-12s %-25s %-16s %10s %s\n", "DATE", "PROJECT", "CATEGORY", "AMOUNT", "NOTES")
	fmt.Println(strings.Repeat("-", 80))
	for _, exp := range expenses {
		spentDate := exp.SpentDate
		if len(spentDate) > 10 {
			spentDate = spentDate[:10]
		}
		amount := fmt.Sprintf("$%.2f", exp.TotalCost)
		projectName := exp.Project.Name
		if len(projectName) > 23 {
			projectName = projectName[:20] + "..."
		}
		categoryName := exp.ExpenseCategory.Name
		if len(categoryName) > 14 {
			categoryName = categoryName[:11] + "..."
		}
		notes := ""
		if exp.Notes != nil {
			notes = *exp.Notes
			if len(notes) > 30 {
				notes = notes[:27] + "..."
			}
		}
		fmt.Printf("%-12s %-25s %-16s %10s %s\n",
			spentDate, projectName, categoryName, amount, notes)
	}
}

func handleExpenseCreate(client *harvest.Client, userIDStr string, logger *log.Logger, projectIDStr, categoryIDStr, amountStr, dateStr, notes string) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Non-interactive mode: all required fields provided
	if projectIDStr != "" && categoryIDStr != "" && amountStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			logger.Fatalf("Invalid project ID: %v", err)
			os.Exit(1)
		}
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			logger.Fatalf("Invalid category ID: %v", err)
			os.Exit(1)
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			logger.Fatalf("Invalid amount: %v", err)
			os.Exit(1)
		}

		req := harvest.ExpenseCreateRequest{
			ProjectID:         projectID,
			ExpenseCategoryID: categoryID,
			SpentDate:         dateStr,
			TotalCost:         amount,
		}
		if notes != "" {
			req.Notes = &notes
		}

		exp, err := client.CreateExpense(req)
		if err != nil {
			logger.Fatalf("Failed to create expense: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Created expense #%d: $%.2f - %s (%s) [%s]\n",
			exp.ID, exp.TotalCost, exp.ExpenseCategory.Name, exp.Project.Name, exp.SpentDate)
		return
	}

	// Interactive mode: select project
	projects, err := client.ListProjects()
	if err != nil {
		logger.Fatalf("Failed to list projects: %v", err)
		os.Exit(1)
	}
	projectOptions := make([]string, len(projects))
	for i, p := range projects {
		projectOptions[i] = fmt.Sprintf("%s \033[36m(%s)\033[0m", p.Name, p.Client.Name)
	}
	idx, err := prompt.SelectPrompt(projectOptions, "Select a project:")
	if err != nil {
		logger.Fatalf("prompt error: %v", err)
		os.Exit(1)
	}
	selectedProject := projects[idx]

	// Interactive mode: select expense category
	categories, err := client.ListExpenseCategories()
	if err != nil {
		logger.Fatalf("Failed to list expense categories: %v", err)
		os.Exit(1)
	}
	if len(categories) == 0 {
		logger.Fatalf("No expense categories found")
		os.Exit(1)
	}
	categoryOptions := make([]string, len(categories))
	for i, c := range categories {
		categoryOptions[i] = c.Name
	}
	idx, err = prompt.SelectPrompt(categoryOptions, "Select an expense category:")
	if err != nil {
		logger.Fatalf("prompt error: %v", err)
		os.Exit(1)
	}
	selectedCategory := categories[idx]

	// Interactive mode: prompt for amount
	amountStr, err = prompt.InputPrompt("Amount:", "")
	if err != nil {
		logger.Fatalf("prompt error: %v", err)
		os.Exit(1)
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logger.Fatalf("Invalid amount: %v", err)
		os.Exit(1)
	}

	// Interactive mode: prompt for notes
	var expenseNotes string
	expenseNotes, err = prompt.InputPrompt("Notes (optional):", notes)
	if err != nil {
		logger.Fatalf("prompt error: %v", err)
		os.Exit(1)
	}

	req := harvest.ExpenseCreateRequest{
		ProjectID:         selectedProject.ID,
		ExpenseCategoryID: selectedCategory.ID,
		SpentDate:         dateStr,
		TotalCost:         amount,
	}
	if expenseNotes != "" {
		req.Notes = &expenseNotes
	}

	exp, err := client.CreateExpense(req)
	if err != nil {
		logger.Fatalf("Failed to create expense: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Created expense #%d: $%.2f - %s (%s) [%s]\n",
		exp.ID, exp.TotalCost, exp.ExpenseCategory.Name, exp.Project.Name, exp.SpentDate)
}

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
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
	var waybarMode bool
	var stopTimer bool
	var addMinutes int
	var lazyProjectSelect bool
	var listInvoices bool
	var listExpenses bool
	var fromDate string
	var toDate string
	var jsonOutput bool
	var createExpense bool
	var expenseProjectID string
	var expenseCategoryID string
	var expenseAmount string
	var expenseDate string
	flag.StringVar(&note, "n", "", "Initial notes text")
	flag.StringVar(&configPath, "c", config.DefaultConfigPath(), "Config file path")
	flag.BoolVar(&ignoreConfig, "i", false, "Ignore loading local configuration")
	flag.BoolVar(&selectEntry, "e", false, "Select and restart an existing time entry")
	flag.BoolVar(&showStatus, "s", false, "Show current running timer status")
	flag.BoolVar(&sketchyBarMode, "b", false, "Format output for SketchyBar (plain text, must be used with -s)")
	flag.BoolVar(&waybarMode, "w", false, "Format output for Waybar (JSON format, must be used with -s)")
	flag.BoolVar(&stopTimer, "q", false, "Stop the currently running timer")
	flag.IntVar(&addMinutes, "a", 0, "Add minutes to current running timer")
	flag.BoolVar(&lazyProjectSelect, "l", false, "Lazy project selection (hide list until typing)")
	flag.BoolVar(&listInvoices, "I", false, "List recent invoices")
	flag.BoolVar(&listExpenses, "E", false, "List expenses")
	flag.StringVar(&fromDate, "from", "", "From date (YYYY-MM-DD)")
	flag.StringVar(&toDate, "to", "", "To date (YYYY-MM-DD)")
	flag.BoolVar(&jsonOutput, "json", false, "Output as raw JSON")
	flag.BoolVar(&createExpense, "create", false, "Create a new expense (must be used with -E)")
	flag.StringVar(&expenseProjectID, "project-id", "", "Project ID for expense")
	flag.StringVar(&expenseCategoryID, "category-id", "", "Expense category ID")
	flag.StringVar(&expenseAmount, "amount", "", "Expense total cost")
	flag.StringVar(&expenseDate, "date", "", "Expense date (YYYY-MM-DD, default: today)")
	var ticket string
	flag.StringVar(&ticket, "t", "", "External ticket number to prefix notes")
	flag.Parse()

	// Validate flags
	if sketchyBarMode && !showStatus {
		logger.Fatalf("-b flag must be used with -s flag")
		os.Exit(1)
	}
	if waybarMode && !showStatus {
		logger.Fatalf("-w flag must be used with -s flag")
		os.Exit(1)
	}
	if sketchyBarMode && waybarMode {
		logger.Fatalf("-b and -w flags cannot be used together")
		os.Exit(1)
	}

	// Validate add minutes flag
	if addMinutes < 0 {
		logger.Fatalf("-a flag must be a positive number of minutes")
		os.Exit(1)
	}

	// Check for conflicting flags with -a
	if addMinutes > 0 && (selectEntry || showStatus || stopTimer) {
		logger.Fatalf("-a flag cannot be used with -e, -s, or -q flags")
		os.Exit(1)
	}

	// Validate --create flag
	if createExpense && !listExpenses {
		logger.Fatalf("--create flag must be used with -E flag")
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
		handleStatusDisplay(client, globalCfg.HarvestUserID, logger, sketchyBarMode, waybarMode)
		return
	}

	// Handle add time mode
	if addMinutes > 0 {
		handleAddTime(client, globalCfg.HarvestUserID, logger, addMinutes)
		return
	}

	// Handle invoice listing
	if listInvoices {
		var from, to *string
		if fromDate != "" {
			from = &fromDate
		}
		if toDate != "" {
			to = &toDate
		}
		handleInvoiceList(client, logger, from, to, jsonOutput)
		return
	}

	// Handle expense listing / creation
	if listExpenses {
		if createExpense {
			handleExpenseCreate(client, globalCfg.HarvestUserID, logger, expenseProjectID, expenseCategoryID, expenseAmount, expenseDate, note)
			return
		}
		var from, to *string
		if fromDate != "" {
			from = &fromDate
		}
		if toDate != "" {
			to = &toDate
		}
		handleExpenseList(client, logger, from, to, jsonOutput)
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
		idx, err := prompt.SelectPromptWithOptions(projectOptions, "Select a project:", lazyProjectSelect)
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
