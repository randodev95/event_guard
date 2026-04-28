package cmd

import (
	"fmt"
	"os"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/spf13/cobra"
)

var oldPlanPath string
var newPlanPath string

// NewDiffCmd initializes the Diff command.
func NewDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two tracking plans for breaking changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load Old Plan
			oldData, err := os.ReadFile(oldPlanPath)
			if err != nil {
				return fmt.Errorf("failed to read old plan: %w", err)
			}
			oldPlan, err := parser.ParseYAML(oldData)
			if err != nil {
				return fmt.Errorf("failed to parse old plan: %w", err)
			}

			// 2. Load New Plan
			newData, err := os.ReadFile(newPlanPath)
			if err != nil {
				return fmt.Errorf("failed to read new plan: %w", err)
			}
			newPlan, err := parser.ParseYAML(newData)
			if err != nil {
				return fmt.Errorf("failed to parse new plan: %w", err)
			}

			// 3. Diff
			breaches := ast.DiffPlans(oldPlan, newPlan)

			if len(breaches) > 0 {
				cmd.Printf("Breaking changes detected between versions [%s] and [%s]:\n", oldPlan.Version, newPlan.Version)
				for _, b := range breaches {
					cmd.Printf("  - %s\n", b)
				}
				os.Exit(1)
			}

			cmd.Printf("No breaking changes detected between [%s] and [%s].\n", oldPlan.Version, newPlan.Version)
			return nil
		},
	}

	cmd.Flags().StringVar(&oldPlanPath, "old", "", "Path to the old tracking plan")
	cmd.Flags().StringVar(&newPlanPath, "new", "canvas.yaml", "Path to the new tracking plan")
	cmd.MarkFlagRequired("old")

	return cmd
}
