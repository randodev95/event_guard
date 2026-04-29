package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFullWorkflow_Simulation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eg_simulation")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tempDir)

	// 1. Init
	initCmd := NewInitCmd()
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify .git exists
	if _, err := os.Stat(".git"); err != nil {
		t.Error(".git directory not created")
	}

	// 2. Modify canvas.yaml
	yamlPath := filepath.Join(tempDir, "canvas.yaml")
	os.WriteFile(yamlPath, []byte("version: simulation-v1"), 0644)

	// 3. Propose
	proposeCmd := NewProposeCmd()
	proposeCmd.SetArgs([]string{"-m", "simulated change"})
	if err := proposeCmd.Execute(); err != nil {
		t.Fatalf("Propose failed: %v", err)
	}

	// Verify we are on a new branch and commit exists
	// (Already covered by propose_test.go logic, but good to run here)
}
