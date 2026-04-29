package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

var proposeMessage string

// NewProposeCmd initializes the Propose command.
func NewProposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Automate branching, committing, and pushing for tracking plan changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Open Repo
			repo, err := git.PlainOpen(".")
			if err != nil {
				return fmt.Errorf("failed to open git repository: %w", err)
			}

			w, err := repo.Worktree()
			if err != nil {
				return err
			}

			// 2. Read canvas.yaml content to memory
			planData, err := os.ReadFile("canvas.yaml")
			if err != nil {
				return fmt.Errorf("failed to read canvas.yaml: %w", err)
			}

			// 3. Create Branch Reference
			head, err := repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get HEAD: %w", err)
			}

			branchName := fmt.Sprintf("analyst/change-%d", time.Now().Unix())
			refName := plumbing.NewBranchReferenceName(branchName)
			cmd.Printf(" Creating branch [%s]...\n", branchName)
			
			ref := plumbing.NewHashReference(refName, head.Hash())
			err = repo.Storer.SetReference(ref)
			if err != nil {
				return fmt.Errorf("failed to create branch reference: %w", err)
			}

			// 4. Switch to it (Force: true ensures we switch even if dirty, we'll restore content)
			err = w.Checkout(&git.CheckoutOptions{
				Branch: refName,
				Force:  true,
			})
			if err != nil {
				return fmt.Errorf("failed to switch to branch: %w", err)
			}

			// 5. Restore canvas.yaml and Stage
			err = os.WriteFile("canvas.yaml", planData, 0644)
			if err != nil {
				return fmt.Errorf("failed to restore canvas.yaml: %w", err)
			}

			cmd.Printf(" Staging canvas.yaml...\n")
			_, err = w.Add("canvas.yaml")
			if err != nil {
				return fmt.Errorf("failed to stage canvas.yaml: %w", err)
			}

			// 6. Commit
			if proposeMessage == "" {
				proposeMessage = "Tracking plan update: " + time.Now().Format(time.RFC3339)
			}
			
			// Try to get author from git config
			var author *object.Signature
			cfg, _ := repo.Config()
			if cfg.User.Name != "" && cfg.User.Email != "" {
				author = &object.Signature{
					Name:  cfg.User.Name,
					Email: cfg.User.Email,
					When:  time.Now(),
				}
			} else {
				author = &object.Signature{
					Name:  "EventGuard Bot",
					Email: "bot@eventguard.io",
					When:  time.Now(),
				}
			}

			cmd.Printf(" Committing changes: %s\n", proposeMessage)
			hash, err := w.Commit(proposeMessage, &git.CommitOptions{
				Author: author,
			})
			if err != nil {
				return fmt.Errorf("commit failed: %w", err)
			}
			cmd.Printf(" Commit created: %s\n", hash.String())

			// 5. Push
			cmd.Printf(" Pushing to remote...\n")
			err = repo.Push(&git.PushOptions{})
			if err != nil {
				if err == git.NoErrAlreadyUpToDate {
					cmd.Println(" Already up to date.")
				} else if err == git.ErrRemoteNotFound {
					cmd.Println(" Warning: No remote found. Skipping push.")
				} else {
					cmd.Printf(" Warning: Push failed (%v). You may need to push manually.\n", err)
				}
			} else {
				cmd.Printf(" Successfully pushed to remote!\n")
			}

			cmd.Printf("\n Propose complete! Now open a Pull Request to merge [%s] into [main].\n", branchName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&proposeMessage, "message", "m", "", "Commit message for the change")

	return cmd
}
