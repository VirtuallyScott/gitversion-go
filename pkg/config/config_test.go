package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Default config when no file provided", func(t *testing.T) {
		config, err := LoadConfig("")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if config.NextVersion != "0.0.0" {
			t.Errorf("Expected NextVersion to be '0.0.0', got '%s'", config.NextVersion)
		}

		if len(config.Branches) == 0 {
			t.Errorf("Expected branches to be configured")
		}
	})

	t.Run("Load JSON config", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test.json")

		jsonContent := `{
			"next-version": "1.0.0",
			"branches": {
				"main": {
					"increment": "Patch",
					"tag": "stable"
				}
			}
		}`

		err := os.WriteFile(configFile, []byte(jsonContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadConfig(configFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if config.NextVersion != "1.0.0" {
			t.Errorf("Expected NextVersion to be '1.0.0', got '%s'", config.NextVersion)
		}

		mainConfig := config.Branches["main"]
		if mainConfig.Increment != "Patch" {
			t.Errorf("Expected main.increment to be 'Patch', got '%s'", mainConfig.Increment)
		}
		if mainConfig.Tag != "stable" {
			t.Errorf("Expected main.tag to be 'stable', got '%s'", mainConfig.Tag)
		}
	})

	t.Run("Load YAML config", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test.yml")

		yamlContent := `next-version: '2.0.0'
branches:
  main:
    increment: Minor
    tag: release
  develop:
    increment: Major
    tag: dev`

		err := os.WriteFile(configFile, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadConfig(configFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if config.NextVersion != "2.0.0" {
			t.Errorf("Expected NextVersion to be '2.0.0', got '%s'", config.NextVersion)
		}

		mainConfig := config.Branches["main"]
		if mainConfig.Increment != "Minor" {
			t.Errorf("Expected main.increment to be 'Minor', got '%s'", mainConfig.Increment)
		}

		developConfig := config.Branches["develop"]
		if developConfig.Increment != "Major" {
			t.Errorf("Expected develop.increment to be 'Major', got '%s'", developConfig.Increment)
		}
	})

	t.Run("Error on non-existent file", func(t *testing.T) {
		_, err := LoadConfig("/non/existent/file.json")
		if err == nil {
			t.Errorf("Expected error for non-existent file")
		}
	})

	t.Run("Error on unsupported format", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test.txt")

		err := os.WriteFile(configFile, []byte("invalid"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err = LoadConfig(configFile)
		if err == nil {
			t.Errorf("Expected error for unsupported format")
		}
	})
}

func TestGetBranchConfig(t *testing.T) {
	config := getDefaultConfig()

	t.Run("Get exact branch config", func(t *testing.T) {
		branchConfig := config.GetBranchConfig("main")
		if branchConfig.Increment != "Patch" {
			t.Errorf("Expected main increment to be 'Patch', got '%s'", branchConfig.Increment)
		}
	})

	t.Run("Get feature branch config", func(t *testing.T) {
		branchConfig := config.GetBranchConfig("feature/user-auth")
		if branchConfig.Increment != "Minor" {
			t.Errorf("Expected feature increment to be 'Minor', got '%s'", branchConfig.Increment)
		}
	})

	t.Run("Get unknown branch config", func(t *testing.T) {
		branchConfig := config.GetBranchConfig("unknown-branch")
		if branchConfig.Increment != "Patch" {
			t.Errorf("Expected unknown increment to be 'Patch', got '%s'", branchConfig.Increment)
		}
		if branchConfig.Tag != "{BranchName}" {
			t.Errorf("Expected unknown tag to be '{BranchName}', got '%s'", branchConfig.Tag)
		}
	})
}
