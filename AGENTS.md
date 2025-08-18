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
