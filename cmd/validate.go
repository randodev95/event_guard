package cmd

import (
	"fmt"
	"os"

	"github.com/randodev95/event_guard/pkg/normalization"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/randodev95/event_guard/pkg/validator"
	"github.com/spf13/cobra"
)

// NewValidateCmd initializes the Validate command.
func NewValidateCmd() *cobra.Command {
	var validatePlanPath string
	var eventNameOverride string

	cmd := &cobra.Command{
		Use:   "validate <file.json>",
		Short: "Validate a JSON payload against the tracking plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load Plan
			plan, err := parser.LoadPlan(validatePlanPath)
			if err != nil {
				return fmt.Errorf("failed to load plan from %s: %w", validatePlanPath, err)
			}

			// 2. Load Payload
			payload, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read payload: %w", err)
			}

			// 3. Normalize
			mapper := normalization.NewDefaultMapper()
			normalized, err := mapper.Map(payload)
			if err != nil {
				return fmt.Errorf("normalization failed: %w", err)
			}

			targetEvent := normalized.Event
			if eventNameOverride != "" {
				targetEvent = eventNameOverride
			}

			// 4. Resolve Schema
			schema, err := plan.ResolveEventSchema(targetEvent)
			if err != nil {
				return err
			}

			// 5. Validate
			result, err := validator.Validate(normalized, schema)
			if err != nil {
				return err
			}

			if result.Valid {
				cmd.Printf(" Event [%s] is VALID\n", targetEvent)
			} else {
				cmd.Printf(" Event [%s] is INVALID:\n", targetEvent)
				for _, e := range result.Errors {
					cmd.Printf("  - %s\n", e)
				}
				return fmt.Errorf("validation failed for event [%s]", targetEvent)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&validatePlanPath, "plan", "p", "maps/", "Path to tracking plan")
	cmd.Flags().StringVarP(&eventNameOverride, "event", "e", "", "Override event name from payload")

	return cmd
}
