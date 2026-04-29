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
			// 1. Create maps/
			if err := os.MkdirAll("maps", 0755); err != nil {
				return err
			}
			defaultYAML := `version: "1.0.0"

taxonomy:
  identity_properties: ["userId"]
  events:
    "Login":
      properties:
        userId: { type: string, required: true }

    "Order Completed":
      properties:
        userId: { type: string, required: true }
        total: { type: number, required: true }

flows:
  Main_Flow:
    namespace: "Main"
    nodes:
      Start:
        type: TriggerNode
        event: "Login"
        transitions: [{target: Purchase}]
      
      Purchase:
        type: WaitNode
        listen_for: "Order Completed"
        transitions: [{target: End}]

      End:
        type: TerminalNode
`
			if err := os.WriteFile(filepath.Join("maps", "plan.yaml"), []byte(defaultYAML), 0644); err != nil {
				return err
			}

			// 2. Create .event_guard directory and event_guard.db
			if err := os.MkdirAll(".event_guard", 0755); err != nil {
				return err
			}

			db, err := storage.NewSQLiteDB(filepath.Join(".event_guard", "event_guard.db"))
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
    paths: [maps/]
jobs:
  impact-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go build -o event_guard main.go
      - run: ./event_guard impact-check --prev-sha ${{ github.event.pull_request.base.sha }}
`
			if err := os.WriteFile(filepath.Join(workflowPath, "event_guard.yml"), []byte(workflowYAML), 0644); err != nil {
				return err
			}

			// 4. Create .gitignore
			gitignore := ".event_guard/\nevent_guard\nevent_guard\n"
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
				w.Add("maps/")
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
