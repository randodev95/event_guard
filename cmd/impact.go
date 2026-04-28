package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/randodev95/event_guard/internal/git"
	"github.com/randodev95/event_guard/internal/impact"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/randodev95/event_guard/internal/storage"
	"github.com/spf13/cobra"
)

var prevSHAFlag string

func NewImpactCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impact-check",
		Short: "Verify schema changes against previous success samples",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 0. Resolve SHA
			targetSHA := prevSHAFlag
			if targetSHA == "" {
				sha, err := git.GetParentSHA(".")
				if err != nil {
					return err
				}
				targetSHA = sha
			}

			// 1. Load Plan
			data, err := os.ReadFile("canvas.yaml")
			if err != nil {
				return fmt.Errorf("failed to read canvas.yaml: %w", err)
			}
			plan, err := parser.ParseYAML(data)
			if err != nil {
				return fmt.Errorf("failed to parse canvas.yaml: %w", err)
			}

			// 2. Load DB
			dbPath := filepath.Join(".canvas", "canvas.db")
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				return fmt.Errorf("local database not found at %s. Run 'canvas init' or 'canvas dev' first", dbPath)
			}

			db, err := storage.NewSQLiteDB(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			// 3. Run Check
			breaches, err := impact.CheckParity(db, targetSHA, plan)
			if err != nil {
				return fmt.Errorf("impact check failed: %w", err)
			}

			if len(breaches) > 0 {
				cmd.Printf("Breaking changes detected against SHA [%s]:\n", targetSHA)
				for _, b := range breaches {
					cmd.Printf("  - Event [%s]: %v\n", b.EventName, b.Errors)
				}
				os.Exit(1)
			}

			cmd.Printf("No breaking changes detected against SHA [%s].\n", targetSHA)
			return nil
		},
	}

	cmd.Flags().StringVar(&prevSHAFlag, "prev-sha", "", "The Git SHA to compare against (default: parent commit)")

	return cmd
}
