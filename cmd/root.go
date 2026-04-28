package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd initializes the base command for EventCanvas.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "canvas",
		Short:   "EventCanvas: Local-first telemetry CLI",
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
	rootCmd.AddCommand(NewServeCmd())

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return NewRootCmd().Execute()
}
