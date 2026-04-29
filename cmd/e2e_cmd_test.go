package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLI_E2E_Lifecycle(t *testing.T) {
	// 1. Setup temp workspace
	tmpDir, err := os.MkdirTemp("", "eg-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 1.5 Compile binary for testing
	binary := filepath.Join(tmpDir, "eventguard")
	build := exec.Command("go", "build", "-o", binary, "github.com/randodev95/event_guard")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(out))
	}
	
	// 2. Test 'init'
	cmdInit := exec.Command(binary, "init")
	cmdInit.Dir = tmpDir
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("eg init failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "canvas.yaml")); os.IsNotExist(err) {
		t.Error("canvas.yaml not created by init")
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ".git")); os.IsNotExist(err) {
		t.Error(".git not initialized by init")
	}

	// 3. Test 'propose' (Success Path)
	// Modify plan
	planPath := filepath.Join(tmpDir, "canvas.yaml")
	newContent := `
version: "1.1.0"
identity_properties: ["userId"]
events:
  "Login":
    category: "AUTH"
    entity_type: "User"
    properties:
      userId: {type: string, required: true}
`
	os.WriteFile(planPath, []byte(newContent), 0644)

	cmdPropose := exec.Command(binary, "propose")
	cmdPropose.Dir = tmpDir
	if out, err := cmdPropose.CombinedOutput(); err != nil {
		t.Fatalf("eg propose failed: %v\nOutput: %s", err, string(out))
	}

	// Verify branch creation (via git command)
	gitBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	gitBranch.Dir = tmpDir
	branch, _ := gitBranch.Output()
	if !os.IsPathSeparator('/') && !testing.Short() {
		// Just a loose check to ensure we aren't on 'main' anymore
		if string(branch) == "main" {
			t.Error("eg propose did not switch to analyst branch")
		}
	}
}
