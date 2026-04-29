package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestPropose_Logic(t *testing.T) {
	// 1. Setup temp repo
	tempDir, err := os.MkdirTemp("", "event_guard_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Add canvas.yaml
	yamlPath := filepath.Join(tempDir, "canvas.yaml")
	if err := os.WriteFile(yamlPath, []byte("version: 1.0.0"), 0644) ; err != nil {
		t.Fatal(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	worktree.Add("canvas.yaml")
	sig := &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()}
	worktree.Commit("initial", &git.CommitOptions{Author: sig, Committer: sig})

	// 3. Modify canvas.yaml
	if err := os.WriteFile(yamlPath, []byte("version: 1.1.0"), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Run Propose
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tempDir)

	proposeCmd := NewProposeCmd()
	proposeCmd.SetArgs([]string{"-m", "test change"})
	if err := proposeCmd.Execute(); err != nil {
		t.Fatalf("propose failed: %v", err)
	}

	// 5. Verify
	head, err := repo.Head()
	if err != nil {
		t.Fatal(err)
	}

	if !head.Name().IsBranch() || !strings.HasPrefix(head.Name().Short(), "analyst/change-") {
		t.Errorf("Expected branch starting with analyst/change-, got %s", head.Name())
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		t.Fatal(err)
	}

	if commit.Message != "test change" {
		t.Errorf("Expected commit message 'test change', got '%s'", commit.Message)
	}
}
