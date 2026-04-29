package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/spf13/cobra"
)

// NewRootCmd initializes the base command for EventGuard.
func NewRootCmd() *cobra.Command {
	var planPath string
	rootCmd := &cobra.Command{
		Use:     "event_guard",
		Short:   "EventGuard: Local-first telemetry CLI",
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&planPath, "plan", "p", "maps", "Path to tracking plan (directory or file)")

	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewDevCmd(&planPath))
	rootCmd.AddCommand(NewDiffCmd())
	rootCmd.AddCommand(NewImpactCheckCmd(&planPath))
	rootCmd.AddCommand(NewGenerateCmd(&planPath))
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewProposeCmd(&planPath))
	rootCmd.AddCommand(NewServeCmd(&planPath))
	rootCmd.AddCommand(NewExportWASMCmd(&planPath))

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return NewRootCmd().Execute()
}

func NewExportWASMCmd(planPath *string) *cobra.Command {
	var hashPlan bool
	cmd := &cobra.Command{
		Use:   "export-wasm",
		Short: "Export the validator as a WASM binary for browser use",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Compiling validator to WASM...")

			// 1. Prepare Plan (Obfuscate if requested)
			plan, err := parser.LoadPlan(*planPath)
			if err != nil {
				return fmt.Errorf("failed to load plan: %w", err)
			}

			if hashPlan {
				cmd.Println(" [SECURITY] Obfuscating plan (hashing names)...")
				_ = plan.Obfuscate()
			}

			buildCmd := exec.Command("go", "build", "-o", "validator.wasm", "web/validator/main.go")
			buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
			
			output, err := buildCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("wasm build failed: %v\n%s", err, string(output))
			}
			
			cmd.Println("Successfully exported [validator.wasm]")
			if hashPlan {
				cmd.Println(" !!! WARNING: Frontend SDK must hash event names before calling egValidate.")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&hashPlan, "hash", false, "Hash event and context names to protect business logic")
	return cmd
}
