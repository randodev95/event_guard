package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/randodev95/event_guard/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_Version(t *testing.T) {
	rootCmd := NewRootCmd()
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"--version"})

	require.NoError(t, rootCmd.Execute())
	assert.Contains(t, b.String(), "event_guard version 0.1.0")
}

func TestInitCommand(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldCwd)

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.SetArgs([]string{"init"})

	require.NoError(t, rootCmd.Execute())

	assert.DirExists(t, "maps")
	assert.DirExists(t, ".event_guard")
	assert.DirExists(t, ".git")
}

func TestProposeCommand(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldCwd)

	// 1. Setup repo
	repo, err := git.PlainInit(".", false)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll("maps", 0755))
	yamlPath := filepath.Join("maps", "plan.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("version: 1.0.0"), 0644))

	w, err := repo.Worktree()
	require.NoError(t, err)
	_, err = w.Add("maps")
	require.NoError(t, err)
	
	sig := &object.Signature{Name: "Test", Email: "test@eg.io", When: time.Now()}
	_, err = w.Commit("initial", &git.CommitOptions{Author: sig, Committer: sig})
	require.NoError(t, err)

	// 2. Modify and Propose
	// NOTE: We modify the file, then call propose. 
	// Propose should stage and commit this change on a new branch.
	require.NoError(t, os.WriteFile(yamlPath, []byte("version: 1.1.0"), 0644))

	root := NewRootCmd()
	// We use the full command path to ensure flags are parsed correctly
	root.SetArgs([]string{"propose", "-m", "test change"})
	
	// We capture output to debug if it fails
	var out bytes.Buffer
	root.SetOut(&out)
	err = root.Execute()
	if err != nil {
		t.Logf("Output: %s", out.String())
	}
	require.NoError(t, err)

	// 3. Verify
	head, err := repo.Head()
	require.NoError(t, err)
	
	// The propose command should have switched us to the new branch
	assert.True(t, strings.HasPrefix(head.Name().Short(), "analyst/change-"), "Should be on analyst branch, got %s", head.Name().Short())
	
	commit, err := repo.CommitObject(head.Hash())
	require.NoError(t, err)
	assert.Equal(t, "test change", commit.Message)
}

func TestImpactCheckCommand(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldCwd)

	require.NoError(t, NewInitCmd().Execute())

	targetSHA := "test-sha"
	dbPath := filepath.Join(".event_guard", "event_guard.db")
	db, err := storage.NewSQLiteDB(dbPath)
	require.NoError(t, err)
	
	require.NoError(t, db.SaveSnapshot(storage.Snapshot{
		SHA:       targetSHA,
		EventName: "Login",
		Payloads:  []string{`{"event": "Login", "userId": "u1"}`},
	}))
	db.Close()

	// Pass case
	root := NewRootCmd()
	root.SetArgs([]string{"impact-check", "--prev-sha", targetSHA})
	require.NoError(t, root.Execute())

	// Breaking change case
	planPath := filepath.Join("maps", "plan.yaml")
	planData, err := os.ReadFile(planPath)
	require.NoError(t, err)
	newPlan := strings.Replace(string(planData), "userId: { type: string, required: true }", "userId: { type: number, required: true }", 1)
	require.NoError(t, os.WriteFile(planPath, []byte(newPlan), 0644))

	root = NewRootCmd()
	root.SetArgs([]string{"impact-check", "--prev-sha", targetSHA})
	require.Error(t, root.Execute())
}
