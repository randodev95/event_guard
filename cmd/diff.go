package cmd

import (
	"fmt"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/spf13/cobra"
)

// NewDiffCmd initializes the Diff command.
func NewDiffCmd() *cobra.Command {
	var oldPlanPath string
	var newPlanPath string
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two tracking plans for breaking changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load Old Plan
			oldPlan, err := parser.LoadPlan(oldPlanPath)
			if err != nil {
				return fmt.Errorf("failed to load old plan: %w", err)
			}

			// 2. Load New Plan
			newPlan, err := parser.LoadPlan(newPlanPath)
			if err != nil {
				return fmt.Errorf("failed to load new plan: %w", err)
			}

			// 3. Diff
			breaches := ast.DiffPlans(oldPlan, newPlan)

			if len(breaches) > 0 {
				cmd.Printf("Breaking changes detected between versions [%s] and [%s]:\n", oldPlan.Version, newPlan.Version)
				for _, b := range breaches {
					cmd.Printf("  - %s\n", b)
				}
				return fmt.Errorf("diff failed with %d breaking changes", len(breaches))
			}

			cmd.Printf("No breaking changes detected between [%s] and [%s].\n", oldPlan.Version, newPlan.Version)
			return nil
		},
	}

	cmd.Flags().StringVar(&oldPlanPath, "old", "", "Path to the old tracking plan")
	cmd.Flags().StringVar(&newPlanPath, "new", "maps/", "Path to the new tracking plan")
	cmd.MarkFlagRequired("old")

	return cmd
}
