package cmd

import (
	"fmt"
	"os"

	"github.com/eventcanvas/eventcanvas/internal/generator"
	"github.com/eventcanvas/eventcanvas/pkg/parser"
	"github.com/spf13/cobra"
)

var target string
var planPath string
var outputPath string

func NewGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate warehouse configurations from the tracking plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load Plan
			data, err := os.ReadFile(planPath)
			if err != nil {
				return fmt.Errorf("failed to read plan: %w", err)
			}
			plan, err := parser.ParseYAML(data)
			if err != nil {
				return fmt.Errorf("failed to parse plan: %w", err)
			}

			// 2. Generate
			var output string
			switch target {
			case "dbt":
				output, err = generator.GenerateDBT(plan)
			case "sqlmesh":
				output, err = generator.GenerateSQLMesh(plan)
			case "html":
				output, err = generator.GenerateHTML(plan)
			case "mermaid":
				output, err = generator.GenerateMermaid(plan)
			default:
				return fmt.Errorf("unsupported target: %s (supported: dbt, sqlmesh, html, mermaid)", target)
			}

			if err != nil {
				return err
			}

			// 3. Output
			if outputPath != "" {
				if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
					return err
				}
				cmd.Printf("Generated %s config to %s\n", target, outputPath)
			} else {
				cmd.Println(output)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "dbt", "Output target (dbt, sqlmesh)")
	cmd.Flags().StringVarP(&planPath, "plan", "p", "canvas.yaml", "Path to tracking plan")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}
