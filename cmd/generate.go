package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/randodev95/event_guard/internal/generator"
	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/spf13/cobra"
)

// NewGenerateCmd initializes the Generate command.
func NewGenerateCmd(planPath *string) *cobra.Command {
	var targets []string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate warehouse configurations from the tracking plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load Plan
			plan, err := parser.LoadPlan(*planPath)
			if err != nil {
				return fmt.Errorf("failed to load plan: %w", err)
			}

			// 2. Parallel Generation
			var wg sync.WaitGroup
			errChan := make(chan error, len(targets))

			for _, t := range targets {
				wg.Add(1)
				go func(targetType string) {
					defer wg.Done()

					output, genErr := runGenerator(targetType, plan)
					if genErr != nil {
						errChan <- fmt.Errorf("target %s failed: %w", targetType, genErr)
						return
					}

					// 3. Output logic
					targetFile := outputPath
					if len(targets) > 1 || outputPath == "" {
						// If multiple targets, we might want to name them specifically
						// For now, if no output path is given, print to stdout with header
						if outputPath == "" {
							cmd.Printf("\n--- Target: %s ---\n%s\n", targetType, output)
							return
						}
						// If output path is given and multiple targets, append suffix
						targetFile = fmt.Sprintf("%s.%s", outputPath, targetType)
					}

					if targetFile != "" {
						if err := os.WriteFile(targetFile, []byte(output), 0644); err != nil {
							errChan <- err
							return
						}
						cmd.Printf("Generated %s config to %s\n", targetType, targetFile)
					}
				}(t)
			}

			wg.Wait()
			close(errChan)

			// Collect errors
			var errors []error
			for e := range errChan {
				errors = append(errors, e)
			}
			if len(errors) > 0 {
				return fmt.Errorf("generation failed with %d errors: %v", len(errors), errors)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&targets, "target", "t", []string{"dbt"}, "Output targets (dbt, sqlmesh, html, mermaid). Can be comma-separated.")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path or prefix (default: stdout)")

	return cmd
}

func runGenerator(targetType string, plan *ast.TrackingPlan) (string, error) {
	switch targetType {
	case "dbt":
		return generator.GenerateDBT(plan)
	case "sqlmesh":
		return generator.GenerateSQLMesh(plan)
	case "html":
		return generator.GenerateHTML(plan)
	case "mermaid":
		return generator.GenerateMermaid(plan)
	default:
		return "", fmt.Errorf("unsupported target: %s", targetType)
	}
}
