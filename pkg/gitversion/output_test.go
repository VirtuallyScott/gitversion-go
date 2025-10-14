package gitversion

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/VirtuallyScott/gitversion-go/pkg/semver"
)

type mockRepo struct{}

func (m *mockRepo) GetSHA() (string, error) {
	return "abc1234567890def", nil
}

func (m *mockRepo) GetShortSHA() (string, error) {
	return "abc1234", nil
}

func (m *mockRepo) GetCommitDate() (string, error) {
	return "2025-01-15 10:30:45 +0000", nil
}

func (m *mockRepo) GetLatestTag() (string, error) {
	return "v1.0.0", nil
}

func (m *mockRepo) GetCommitCountSinceTag(tag string) (int, error) {
	return 5, nil
}

func TestFormat(t *testing.T) {
	formatter := NewFormatter(&mockRepo{})
	version := &semver.Version{
		Major:      1,
		Minor:      2,
		Patch:      3,
		PreRelease: "alpha.5",
		Build:      "10+abc1234",
	}

	tests := []struct {
		name     string
		format   OutputFormat
		expected string
	}{
		{
			name:     "Text format",
			format:   Text,
			expected: "1.2.3-alpha.5+10+abc1234",
		},
		{
			name:     "AssemblySemVer format",
			format:   AssemblySemVer,
			expected: "1.2.3.0",
		},
		{
			name:     "AssemblySemFileVer format",
			format:   AssemblySemFileVer,
			expected: "1.2.3.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(version, tt.format, "develop")
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.format == JSON {
				// For JSON, just check if it's valid JSON
				var jsonOutput JSONOutput
				if err := json.Unmarshal([]byte(result), &jsonOutput); err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}
			} else {
				if result != tt.expected {
					t.Errorf("Format(%s) = %s, want %s", tt.format, result, tt.expected)
				}
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	formatter := NewFormatter(&mockRepo{})
	version := &semver.Version{
		Major:      1,
		Minor:      2,
		Patch:      3,
		PreRelease: "alpha.5",
		Build:      "10+abc1234",
	}

	result, err := formatter.formatJSON(version, "develop")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test specific fields
	if output.Major != 1 {
		t.Errorf("Major = %d, want 1", output.Major)
	}
	if output.Minor != 2 {
		t.Errorf("Minor = %d, want 2", output.Minor)
	}
	if output.Patch != 3 {
		t.Errorf("Patch = %d, want 3", output.Patch)
	}
	if output.PreReleaseTag != "alpha.5" {
		t.Errorf("PreReleaseTag = %s, want alpha.5", output.PreReleaseTag)
	}
	if output.PreReleaseTagWithDash != "-alpha.5" {
		t.Errorf("PreReleaseTagWithDash = %s, want -alpha.5", output.PreReleaseTagWithDash)
	}
	if output.BuildMetaData != "10+abc1234" {
		t.Errorf("BuildMetaData = %s, want 10+abc1234", output.BuildMetaData)
	}
	if output.BuildMetaDataPadded != "+10+abc1234" {
		t.Errorf("BuildMetaDataPadded = %s, want +10+abc1234", output.BuildMetaDataPadded)
	}
	if output.MajorMinorPatch != "1.2.3" {
		t.Errorf("MajorMinorPatch = %s, want 1.2.3", output.MajorMinorPatch)
	}
	if output.SemVer != "1.2.3-alpha.5+10+abc1234" {
		t.Errorf("SemVer = %s, want 1.2.3-alpha.5+10+abc1234", output.SemVer)
	}
	if output.AssemblySemVer != "1.2.3.0" {
		t.Errorf("AssemblySemVer = %s, want 1.2.3.0", output.AssemblySemVer)
	}
	if output.AssemblySemFileVer != "1.2.3.0" {
		t.Errorf("AssemblySemFileVer = %s, want 1.2.3.0", output.AssemblySemFileVer)
	}
	if output.BranchName != "develop" {
		t.Errorf("BranchName = %s, want develop", output.BranchName)
	}
	if output.EscapedBranchName != "develop" {
		t.Errorf("EscapedBranchName = %s, want develop", output.EscapedBranchName)
	}
	if output.Sha != "abc1234567890def" {
		t.Errorf("Sha = %s, want abc1234567890def", output.Sha)
	}
	if output.ShortSha != "abc1234" {
		t.Errorf("ShortSha = %s, want abc1234", output.ShortSha)
	}
	if output.CommitsSinceVersionSource != 5 {
		t.Errorf("CommitsSinceVersionSource = %d, want 5", output.CommitsSinceVersionSource)
	}
	if output.CommitDate != "2025-01-15 10:30:45 +0000" {
		t.Errorf("CommitDate = %s, want 2025-01-15 10:30:45 +0000", output.CommitDate)
	}
}

func TestFormatJSONWithoutPrerelease(t *testing.T) {
	formatter := NewFormatter(&mockRepo{})
	version := &semver.Version{
		Major: 1,
		Minor: 2,
		Patch: 3,
		Build: "5+abc1234",
	}

	result, err := formatter.formatJSON(version, "main")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if output.PreReleaseTag != "" {
		t.Errorf("PreReleaseTag = %s, want empty string", output.PreReleaseTag)
	}
	if output.PreReleaseTagWithDash != "" {
		t.Errorf("PreReleaseTagWithDash = %s, want empty string", output.PreReleaseTagWithDash)
	}
}

func TestFormatInvalidFormat(t *testing.T) {
	formatter := NewFormatter(&mockRepo{})
	version := &semver.Version{Major: 1, Minor: 2, Patch: 3}

	_, err := formatter.Format(version, OutputFormat("invalid"), "main")
	if err == nil {
		t.Errorf("Expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unknown output format") {
		t.Errorf("Error should mention unknown output format, got: %v", err)
	}
}
