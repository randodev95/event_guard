package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/randodev95/event_guard/internal/api"
	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/randodev95/event_guard/pkg/validator"
	"github.com/spf13/cobra"
)

var port int
var servePlanPath string

// NewServeCmd initializes the Serve command.
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the EventCanvas validation server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Senior Pattern: Support Environment Variables (12-Factor App)
			if p := os.Getenv("EVENTCANVAS_PORT"); p != "" {
				if val, err := strconv.Atoi(p); err == nil {
					port = val
				}
			}
			if p := os.Getenv("EVENTCANVAS_PLAN"); p != "" {
				servePlanPath = p
			}

			// 1. Load Plan
			data, err := os.ReadFile(servePlanPath)
			if err != nil {
				return fmt.Errorf("failed to read plan: %w", err)
			}
			plan, err := parser.ParseYAML(data)
			if err != nil {
				return fmt.Errorf("failed to parse plan: %w", err)
			}

			// 2. Initialize Engine & Server
			engine := validator.NewEngine(plan)
			handler := api.NewServer(engine)

			srv := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: handler,
			}

			// 3. Graceful Shutdown Implementation
			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				slog.Info(" EventCanvas server starting",
					"addr", srv.Addr,
					"plan", servePlanPath)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("server failed", "err", err)
					os.Exit(1)
				}
			}()

			<-done
			slog.Info(" Server stopping...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				slog.Error("graceful shutdown failed", "err", err)
				return err
			}

			slog.Info(" Server stopped gracefully")
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "v", 8080, "Port to listen on")
	cmd.Flags().StringVarP(&servePlanPath, "plan", "p", "canvas.yaml", "Path to tracking plan")

	return cmd
}
