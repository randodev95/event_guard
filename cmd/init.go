package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/randodev95/event_guard/internal/storage"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

// NewInitCmd initializes the Init command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new EventGuard project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Create canvas.yaml
			defaultYAML := `
identity_properties: ["userId"]

contexts:
  Universal_User_Context:
    entity_type: "User"
    properties:
      userId: { type: string, required: true }

events:
  "Login Started":
    category: "INTERACTION"
    entity_type: "User"
    inherits: ["Universal_User_Context"]
    triggers:
      - from_state: "Landing"
        type: "DIRECT_LOAD"

  "Order Completed":
    category: "TRANSACTION"
    entity_type: "Transaction"
    inherits: ["Universal_User_Context"]
    properties:
      total: { type: number, required: true }

flows:
  - id: "onboarding_flow"
    name: "User Onboarding"
    steps:
      - state: "Landing"
        event: "Login Started"
        triggers: ["DIRECT_LOAD"]
`
			if err := os.WriteFile("canvas.yaml", []byte(defaultYAML), 0644); err != nil {
				return err
			}

			// 2. Create .canvas directory and canvas.db
			if err := os.MkdirAll(".canvas", 0755); err != nil {
				return err
			}

			db, err := storage.NewSQLiteDB(filepath.Join(".canvas", "canvas.db"))
			if err != nil {
				return err
			}
			db.Close()

			// 3. Create .github/workflows/event_guard.yml
			workflowPath := filepath.Join(".github", "workflows")
			if err := os.MkdirAll(workflowPath, 0755); err != nil {
				return err
			}
			workflowYAML := `name: EventGuard Telemetry Guard
on:
  pull_request:
    paths: [canvas.yaml]
jobs:
  impact-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go build -o canvas main.go
      - run: ./canvas impact-check --prev-sha ${{ github.event.pull_request.base.sha }}
`
			if err := os.WriteFile(filepath.Join(workflowPath, "event_guard.yml"), []byte(workflowYAML), 0644); err != nil {
				return err
			}

			// 4. Create .gitignore
			gitignore := ".canvas/\ncanvas\nevent_guard\n"
			if err := os.WriteFile(".gitignore", []byte(gitignore), 0644); err != nil {
				return err
			}

			// 5. Create Project README
			projectREADME := `# Tracking Plan: EventGuard
This repository contains the deterministic tracking plan for our telemetry pipeline.

## Usage
- **Propose changes**: ` + "`" + `event_guard propose -m "message"` + "`" + `
- **Validate local data**: ` + "`" + `event_guard validate payload.json` + "`" + `
- **Explore flows**: ` + "`" + `event_guard dev` + "`" + `
`
			if err := os.WriteFile("README.md", []byte(projectREADME), 0644); err != nil {
				return err
			}

			// 6. Initialize Git and Genesis Commit
			repo, err := git.PlainInit(".", false)
			if err != nil {
				if err == git.ErrRepositoryAlreadyExists {
					cmd.Println("Git repository already exists, skipping initialization.")
				} else {
					return err
				}
			} else {
				w, _ := repo.Worktree()
				w.Add("canvas.yaml")
				w.Add(".gitignore")
				w.Add("README.md")
				w.Add(".github/workflows/event_guard.yml")

				sig := &object.Signature{
					Name:  "EventGuard Bot",
					Email: "bot@eventguard.io",
					When:  time.Now(),
				}
				
				hash, err := w.Commit("chore: genesis commit (event_guard init)", &git.CommitOptions{
					Author: sig,
				})
				if err != nil {
					return err
				}
				cmd.Printf("Created genesis commit on [main]: %s\n", hash.String())
			}

			cmd.Println("Successfully initialized EventGuard project.")
			return nil
		},
	}
}
