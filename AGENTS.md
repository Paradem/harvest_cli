# Agent Instructions

## Build / Lint / Test
- **Build**: `go build ./...`
- **Lint**: `golangci-lint run` (runs formatting, vet, staticcheck, errcheck, etc.)
- **Run all tests**: `go test ./... -v`
- **Run a single file test**: `go test -run TestXYZ ./internal/...` or specify the package and test name.

## Code Style Guidelines
- Use `gofmt` formatting; run `go fmt ./...` before committing.
- Imports sorted: standard library first, then third‑party, then local packages. No unused imports.
- Prefer typed errors (`errors.New`, `fmt.Errorf`) over panic; use context where appropriate.
- Function names in CamelCase; exported symbols start with a capital letter.
- Variable names short but descriptive; constants in ALL_CAPS.
- Keep functions small (≤ 50 lines) and single‑responsibility.
- Use Go modules; keep `go.mod` up to date.

## Cursor / Copilot Rules
No `.cursor` or `.cursorrules` directories found. No Copilot instructions present.

## Harvest CLI Project Status

### Current Features Implemented

#### CLI Flags
- **Default behavior**: Create new time entry (interactive project/task selection)
- **`-n <notes>`**: Add notes to new time entry
- **`-c <config_path>`**: Specify custom config file path
- **`-i`**: Ignore local configuration file
- **`-t <ticket>`**: Add ticket number prefix (auto-adds # if missing)
- **`-e`**: Select and restart existing time entries for today
- **`-s`**: Show current running timer status

#### Environment Variables Required
- `HARVEST_ACCOUNT_ID`: Harvest API account ID
- `HARVEST_ACCESS_TOKEN`: Harvest API access token
- `HARVEST_USER_ID`: User ID (required for `-e` and `-s` flags)

### Architecture Overview

#### File Structure
```
cmd/root.go              # Main CLI entry point and flag handling
internal/
  config/config.go       # Configuration management
  harvest/
    client.go           # API client and HTTP requests
    models.go           # Data structures for Harvest API
  prompt/prompt.go      # Interactive selection UI (bubbletea)
```

#### Key Components
- **Harvest API Client**: RESTful client for Harvest API v2
- **Interactive Prompts**: Bubbletea-based selection interfaces
- **Configuration System**: JSON-based local config for defaults
- **Time Entry Management**: Create, list, and restart time entries

### Implementation Details

#### Time Entry Display Formatting
- **Project Selection**: `Project Name (Client Name)` with cyan client highlighting
- **Task Selection**: Just task name (no ID)
- **Time Entry List (-e)**: `Project - Task (Status) [Hours] Notes`
  - Status: Green for running, Yellow for stopped
  - Notes: Cyan highlighting, truncated to 60 chars, newlines converted to `|`
- **Status Display (-s)**: `[HH:MM] #first-word-of-notes` or `[xx:xx]` if no timer

#### API Integration
- **List Time Entries**: Supports date filtering and user filtering
- **Restart Timer**: PATCH `/time_entries/{id}/restart`
- **Create Time Entry**: POST `/time_entries` (duration-based)
- **Project/Task Lists**: GET with active filtering

#### Error Handling
- **Missing HARVEST_USER_ID**: Exits with error (required for user-specific operations)
- **API Errors**: Propagated with context
- **Invalid Input**: Validation with clear error messages

### Known Behaviors & Design Decisions

#### User Experience
- **ID-Free Displays**: All selection lists hide IDs for cleaner interface
- **Color Coding**: Status indicators and notes highlighting
- **Smart Defaults**: Remembers last project/task selection
- **Notes Processing**: Handles multi-line notes, extracts first word for status

#### Technical Choices
- **Bubbletea Framework**: For interactive TUI components
- **No External Color Libraries**: Uses ANSI escape codes directly
- **Stateless Design**: Each command is independent
- **Environment-Based Auth**: Follows 12-factor app principles

### Potential Future Enhancements

#### High Priority
- Add tests for all major functions
- Add `--help` flag documentation
- Add validation for environment variables on startup

#### Medium Priority
- Add `--dry-run` flag for testing
- Add `--verbose` flag for detailed output
- Add support for time entry deletion
- Add project/task filtering options

#### Low Priority
- Add configuration validation
- Add rate limiting for API calls
- Add caching for frequently accessed data
- Add support for multiple concurrent timers

### Development Notes

#### Testing the Application
```bash
# Set required environment variables
export HARVEST_ACCOUNT_ID="your_id"
export HARVEST_ACCESS_TOKEN="your_token"
export HARVEST_USER_ID="your_user_id"

# Test different modes
./harvest_cli                    # Create new entry
./harvest_cli -e                 # Select existing entry
./harvest_cli -s                 # Show status
./harvest_cli -n "Working on feature"  # With notes
```

#### Common Issues
- **Missing Environment Variables**: Ensure all three are set
- **API Rate Limits**: Harvest has rate limits, implement backoff if needed
- **Time Zone Handling**: All times are in user's local timezone
- **Timer Conflicts**: Only one running timer supported (Harvest limitation)

#### Code Quality Notes
- All code formatted with `go fmt`
- Imports properly sorted
- Functions kept small and focused
- Error handling consistent throughout
- No unused imports or variables
