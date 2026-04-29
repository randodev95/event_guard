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
	"github.com/fsnotify/fsnotify"
)

// NewServeCmd initializes the Serve command.
func NewServeCmd(planPath *string) *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the EventGuard validation server",
		RunE: func(cmd *cobra.Command, args []string) error {
			currentPlanPath := *planPath

			// Support Environment Variables (12-Factor App)
			if p := os.Getenv("EVENT_GUARD_PORT"); p != "" {
				if val, err := strconv.Atoi(p); err == nil {
					port = val
				}
			}
			if p := os.Getenv("EVENT_GUARD_PLAN"); p != "" {
				currentPlanPath = p
			}

			// 1. Load Plan
			plan, err := parser.LoadPlan(currentPlanPath)
			if err != nil {
				return fmt.Errorf("failed to load plan from %s: %w", currentPlanPath, err)
			}

			// 2. Initialize Engine & Server
			engine := validator.NewEngine(plan)
			engine.Warmup()
			handler := api.NewServer(engine)

			// 3. Configure Sink
			var baseSink api.Sink
			if url := os.Getenv("EVENT_GUARD_DLQ_WEBHOOK"); url != "" {
				baseSink = api.NewWebhookSink(url)
			} else {
				baseSink = api.NewFileSink("dlq.jsonl")
			}
			
			// Wrap in AsyncSink for scale
			sink := api.NewAsyncSink(baseSink, 1000, 5)
			defer sink.Close()
			handler.SetSink(sink)

			// 4. Admin Reload Handler
			handler.SetReloadHandler(func() error {
				newPlan, err := parser.LoadPlan(currentPlanPath)
				if err != nil {
					return err
				}
				newEngine := validator.NewEngine(newPlan)
				newEngine.Warmup()
				handler.UpdateEngine(newEngine)
				return nil
			})

			// 5. Live Reload
			watcher, err := fsnotify.NewWatcher()
			if err == nil {
				go func() {
					for {
						select {
						case event, ok := <-watcher.Events:
							if !ok { return }
							if event.Op&fsnotify.Write == fsnotify.Write {
								slog.Info("config changed, reloading...", "file", currentPlanPath)
								newPlan, err := parser.LoadPlan(currentPlanPath)
								if err == nil {
									newEngine := validator.NewEngine(newPlan)
									newEngine.Warmup()
									handler.UpdateEngine(newEngine)
								}
							}
						case err, ok := <-watcher.Errors:
							if !ok { return }
							slog.Error("watcher error", "err", err)
						}
					}
				}()
				watcher.Add(currentPlanPath)
			}

			srv := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: handler,
			}

			// 3. Graceful Shutdown Implementation
			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				slog.Info(" EventGuard server starting",
					"addr", srv.Addr,
					"plan", currentPlanPath)
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

	return cmd
}
