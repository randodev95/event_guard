package cmd

import (
	"os"
	"path/filepath"
	"github.com/spf13/cobra"
	"github.com/randodev95/event_guard/internal/storage"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new EventCanvas project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Create canvas.yaml
			defaultYAML := `
contexts:
  Universal_User_Context:
    entity_type: "User"
    properties:
      userId: { type: string, required: true }

events:
  "Order Completed":
    category: "INTERACTION"
    entity_type: "Transaction"
    inherits: ["Universal_User_Context"]
    properties:
      total: { type: number, required: true }
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

			// 3. Create .github/workflows/eventcanvas.yml
			workflowPath := filepath.Join(".github", "workflows")
			if err := os.MkdirAll(workflowPath, 0755); err != nil {
				return err
			}
			workflowYAML := `name: EventCanvas Telemetry Guard
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
			if err := os.WriteFile(filepath.Join(workflowPath, "eventcanvas.yml"), []byte(workflowYAML), 0644); err != nil {
				return err
			}

			cmd.Println("Initialized EventCanvas project with GitHub Action.")
			return nil
		},
	}
}
