package cmd

import (
	"fmt"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/randodev95/event_guard/internal/server"
	"github.com/randodev95/event_guard/internal/tui"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/spf13/cobra"
)

// NewDevCmd initializes the Dev command.
func NewDevCmd(planPath *string) *cobra.Command {
	var headless bool
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start the local development mock server and TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load plan
			plan, err := parser.LoadPlan(*planPath)
			if err != nil {
				return fmt.Errorf("failed to load plan: %w", err)
			}

			// 2. Setup EventBus
			updates := make(chan tui.EventMsg)

			// 3. Setup Server
			srv := &server.Server{
				Plan:    plan,
				Updates: updates,
			}

			// Start server in background
			go func() {
				mux := http.NewServeMux()
				mux.HandleFunc("/", srv.HandleEvent)
				mux.HandleFunc("/api/plan", srv.HandlePlan)
				mux.HandleFunc("/api/events", srv.HandleEvents)
				http.ListenAndServe(":8080", mux)
			}()

			if headless {
				cmd.Printf(" Headless mode: Listening for events...\n")
				for up := range updates {
					status := " VALID"
					if !up.IsValid {
						status = " INVALID"
					}
					cmd.Printf("[%s] Event: %s\n", status, up.Name)
					if !up.IsValid {
						for _, e := range up.Errors {
							cmd.Printf("  - %s\n", e)
						}
					}
				}
				return nil
			}

			// 4. Setup TUI
			p := tea.NewProgram(tui.NewDashboard(updates))
			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&headless, "headless", false, "Run server without TUI")
	return cmd
}
