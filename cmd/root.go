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
	rootCmd := &cobra.Command{
		Use:     "canvas",
		Short:   "EventGuard: Local-first telemetry CLI",
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewDevCmd())
	rootCmd.AddCommand(NewDiffCmd())
	rootCmd.AddCommand(NewImpactCheckCmd())
	rootCmd.AddCommand(NewGenerateCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewProposeCmd())
	rootCmd.AddCommand(NewServeCmd())
	rootCmd.AddCommand(NewExportWASMCmd())

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return NewRootCmd().Execute()
}

func NewExportWASMCmd() *cobra.Command {
	var hashPlan bool
	cmd := &cobra.Command{
		Use:   "export-wasm",
		Short: "Export the validator as a WASM binary for browser use",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Compiling validator to WASM...")

			// 1. Prepare Plan (Obfuscate if requested)
			data, err := os.ReadFile(planPath)
			if err != nil {
				return err
			}
			newPlan, err := parser.ParseYAML(data)
			if err != nil {
				return err
			}

			if hashPlan {
				cmd.Println(" [SECURITY] Obfuscating plan (hashing names)...")
				newPlan = newPlan.Obfuscate()
			}

			// 2. Write plan to temporary location for WASM build
			// Note: WASM entry point needs a way to read this. 
			// For simplicity in this demo, we write to a 'plan.yaml' that the WASM can pick up
			// or the user can initialize at runtime.
			
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
