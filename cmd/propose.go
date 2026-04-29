package cmd

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

// NewProposeCmd initializes the Propose command.
func NewProposeCmd(planPath *string) *cobra.Command {
	var proposeMessage string
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

			// 2. Stage changes
			// We stage everything to be safe.
			_, err = w.Add(".")
			if err != nil {
				return fmt.Errorf("failed to stage changes: %w", err)
			}

			// 3. Create Branch Reference
			head, err := repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get HEAD: %w", err)
			}

			branchName := fmt.Sprintf("analyst/change-%d", time.Now().Unix())
			refName := plumbing.NewBranchReferenceName(branchName)
			
			// Point the new branch to current head
			ref := plumbing.NewHashReference(refName, head.Hash())
			err = repo.Storer.SetReference(ref)
			if err != nil {
				return fmt.Errorf("failed to create branch reference: %w", err)
			}

			// 4. Update HEAD to point to the new branch WITHOUT touching the worktree
			// This is equivalent to 'git checkout -b' but safer for uncommitted changes
			err = repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, refName))
			if err != nil {
				return fmt.Errorf("failed to update HEAD: %w", err)
			}

			// 5. Commit
			msg := proposeMessage
			if msg == "" {
				msg = "Tracking plan update: " + time.Now().Format(time.RFC3339)
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

			hash, err := w.Commit(msg, &git.CommitOptions{
				Author: author,
			})
			if err != nil {
				if err == git.ErrEmptyCommit {
					cmd.Println(" No changes detected. Nothing to propose.")
					return nil
				}
				return fmt.Errorf("commit failed: %w", err)
			}
			cmd.Printf(" Commit created: %s\n", hash.String())

			// 6. Push
			err = repo.Push(&git.PushOptions{})
			if err != nil && err != git.NoErrAlreadyUpToDate && err != git.ErrRemoteNotFound {
				cmd.Printf(" Warning: Push failed (%v). You may need to push manually.\n", err)
			}

			cmd.Printf("\n Propose complete! Now open a Pull Request to merge [%s] into [main].\n", branchName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&proposeMessage, "message", "m", "", "Commit message for the change")

	return cmd
}
