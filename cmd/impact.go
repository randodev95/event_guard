package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/randodev95/event_guard/internal/git"
	"github.com/randodev95/event_guard/internal/impact"
	"github.com/randodev95/event_guard/internal/storage"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/spf13/cobra"
)

// NewImpactCheckCmd initializes the ImpactCheck command.
func NewImpactCheckCmd(planPath *string) *cobra.Command {
	var prevSHAFlag string
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
			plan, err := parser.LoadPlan(*planPath)
			if err != nil {
				return fmt.Errorf("failed to load plan: %w", err)
			}

			// 2. Load DB
			dbPath := filepath.Join(".event_guard", "event_guard.db")
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				return fmt.Errorf("local database not found at %s. Run 'event_guard init' or 'event_guard dev' first", dbPath)
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
				return fmt.Errorf("impact check failed with %d breaches", len(breaches))
			}

			cmd.Printf("No breaking changes detected against SHA [%s].\n", targetSHA)
			return nil
		},
	}

	cmd.Flags().StringVar(&prevSHAFlag, "prev-sha", "", "The Git SHA to compare against (default: parent commit)")

	return cmd
}
