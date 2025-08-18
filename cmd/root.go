package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/example/harvestcli/internal/config"
	"github.com/example/harvestcli/internal/harvest"
	"github.com/example/harvestcli/internal/prompt"
)

func main() {
	var note string
	var configPath string
	flag.StringVar(&note, "n", "", "Initial notes text")
	flag.StringVar(&configPath, "c", config.DefaultConfigPath(), "Config file path")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := harvest.NewClient()
	if err != nil {
		log.Fatalf("Auth error: %v", err)
	}

	// Projects selection
	projects, err := client.ListProjects()
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}
	projectOptions := make([]string, len(projects))
	for i, p := range projects {
		projectOptions[i] = fmt.Sprintf("%d: %s", p.ID, p.Name)
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
		taskOptions[i] = fmt.Sprintf("%d: %s", t.ID, t.Name)
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
	if err := cfg.Save(configPath); err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}
