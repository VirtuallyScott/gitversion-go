package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGitVersionCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Build the binary first
	buildDir := t.TempDir()
	binaryPath := filepath.Join(buildDir, "gitversion")

	// Build from project root to ensure go.mod is accessible
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd")
	buildCmd.Dir = projectRoot
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Create a test git repository
	testRepo := t.TempDir()

	// Initialize git repository
	initGit(t, testRepo)

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T, repoDir string)
		validate func(t *testing.T, output string, err error)
	}{
		{
			name:  "Help flag",
			args:  []string{"--help"},
			setup: func(t *testing.T, repoDir string) {},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !strings.Contains(output, "GitVersion Go implementation") {
					t.Errorf("Help output should contain description")
				}
			},
		},
		{
			name:  "Version flag",
			args:  []string{"--version"},
			setup: func(t *testing.T, repoDir string) {},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !strings.Contains(output, "gitversion v") {
					t.Errorf("Version output should contain version info")
				}
			},
		},
		{
			name: "Basic version calculation",
			args: []string{},
			setup: func(t *testing.T, repoDir string) {
				createCommit(t, repoDir, "Initial commit")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				output = strings.TrimSpace(output)
				// With our GitVersion config, should start with 1.0.0 (next-version)
				if !strings.HasPrefix(output, "1.0.0") {
					t.Errorf("Expected version to start with 1.0.0, got: %s", output)
				}
			},
		},
		{
			name: "JSON output format",
			args: []string{"--output", "json"},
			setup: func(t *testing.T, repoDir string) {
				createCommit(t, repoDir, "feat: add new feature")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				var jsonOutput map[string]interface{}
				if err := json.Unmarshal([]byte(output), &jsonOutput); err != nil {
					t.Errorf("Output should be valid JSON: %v", err)
				}

				if _, exists := jsonOutput["Major"]; !exists {
					t.Errorf("JSON output should contain Major field")
				}
				if _, exists := jsonOutput["SemVer"]; !exists {
					t.Errorf("JSON output should contain SemVer field")
				}
			},
		},
		{
			name:  "AssemblySemVer output format",
			args:  []string{"--output", "AssemblySemVer"},
			setup: func(t *testing.T, repoDir string) {},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				output = strings.TrimSpace(output)
				if !strings.HasSuffix(output, ".0") {
					t.Errorf("AssemblySemVer should end with .0, got: %s", output)
				}
			},
		},
		{
			name: "Force major increment",
			args: []string{"--major"},
			setup: func(t *testing.T, repoDir string) {
				createCommit(t, repoDir, "initial commit")
				createTag(t, repoDir, "v0.1.0")
				createCommit(t, repoDir, "fix: minor bug")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				output = strings.TrimSpace(output)
				// Should increment major version from 0.1.0 to 1.0.0
				if !strings.HasPrefix(output, "1.0.0") {
					t.Errorf("Expected major increment to 1.0.0, got: %s", output)
				}
			},
		},
		{
			name: "GitHub Flow workflow",
			args: []string{"--workflow", "githubflow"},
			setup: func(t *testing.T, repoDir string) {
				createBranch(t, repoDir, "feature/test")
				createCommit(t, repoDir, "feat: add test feature")
			},
			validate: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Should work without error for GitHub Flow
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh repository for each test
			repoDir := t.TempDir()
			initGit(t, repoDir)

			// Run test setup
			tt.setup(t, repoDir)

			// Run the gitversion command
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = repoDir
			output, err := cmd.CombinedOutput()

			tt.validate(t, string(output), err)
		})
	}
}

func initGit(t *testing.T, dir string) {
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmd := range commands {
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		execCmd.Dir = dir
		if err := execCmd.Run(); err != nil {
			t.Fatalf("Failed to run %v: %v", cmd, err)
		}
	}

	// Create a simple gitversion.yml for tests
	gitVersionConfig := `---
next-version: 1.0.0
mode: ContinuousDelivery
increment: Inherit

branches:
  main:
    increment: Patch
    mode: ManualDeployment

  master:
    increment: Patch
    mode: ManualDeployment

ignore:
  sha: []`

	configFile := filepath.Join(dir, "gitversion.yml")
	if err := os.WriteFile(configFile, []byte(gitVersionConfig), 0644); err != nil {
		t.Fatalf("Failed to create gitversion.yml: %v", err)
	}
}

func createCommit(t *testing.T, repoDir, message string) {
	// Create a test file with unique content each time
	testFile := filepath.Join(repoDir, "test.txt")
	content := fmt.Sprintf("test content - %s - %d", message, time.Now().UnixNano())
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	commands := [][]string{
		{"git", "add", "test.txt"},
		{"git", "commit", "-m", message},
	}

	for _, cmd := range commands {
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		execCmd.Dir = repoDir
		if err := execCmd.Run(); err != nil {
			t.Fatalf("Failed to run %v: %v", cmd, err)
		}
	}
}

func createTag(t *testing.T, repoDir, tag string) {
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create tag %s: %v", tag, err)
	}
}

func createBranch(t *testing.T, repoDir, branch string) {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch %s: %v", branch, err)
	}
}
