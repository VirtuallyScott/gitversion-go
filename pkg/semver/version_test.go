package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expected    *Version
		expectError bool
	}{
		{
			name:    "Simple version",
			version: "1.2.3",
			expected: &Version{
				Major: 1, Minor: 2, Patch: 3,
			},
		},
		{
			name:    "Version with v prefix",
			version: "v1.2.3",
			expected: &Version{
				Major: 1, Minor: 2, Patch: 3,
			},
		},
		{
			name:    "Version with prerelease",
			version: "1.2.3-alpha.1",
			expected: &Version{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1",
			},
		},
		{
			name:    "Version with build",
			version: "1.2.3+build.1",
			expected: &Version{
				Major: 1, Minor: 2, Patch: 3, Build: "build.1",
			},
		},
		{
			name:    "Version with prerelease and build",
			version: "1.2.3-alpha.1+build.1",
			expected: &Version{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "build.1",
			},
		},
		{
			name:        "Invalid version",
			version:     "invalid",
			expectError: true,
		},
		{
			name:        "Invalid major",
			version:     "x.2.3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.version)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Major != tt.expected.Major {
				t.Errorf("Major mismatch: got %d, want %d", result.Major, tt.expected.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("Minor mismatch: got %d, want %d", result.Minor, tt.expected.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("Patch mismatch: got %d, want %d", result.Patch, tt.expected.Patch)
			}
			if result.PreRelease != tt.expected.PreRelease {
				t.Errorf("PreRelease mismatch: got %s, want %s", result.PreRelease, tt.expected.PreRelease)
			}
			if result.Build != tt.expected.Build {
				t.Errorf("Build mismatch: got %s, want %s", result.Build, tt.expected.Build)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{
			name:     "Simple version",
			version:  &Version{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "Version with prerelease",
			version:  &Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
			expected: "1.2.3-alpha.1",
		},
		{
			name:     "Version with build",
			version:  &Version{Major: 1, Minor: 2, Patch: 3, Build: "build.1"},
			expected: "1.2.3+build.1",
		},
		{
			name:     "Version with prerelease and build",
			version:  &Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "build.1"},
			expected: "1.2.3-alpha.1+build.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("String mismatch: got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestVersionIncrement(t *testing.T) {
	t.Run("IncrementMajor", func(t *testing.T) {
		v := &Version{Major: 1, Minor: 2, Patch: 3}
		v.IncrementMajor()

		if v.Major != 2 || v.Minor != 0 || v.Patch != 0 {
			t.Errorf("IncrementMajor failed: got %d.%d.%d, want 2.0.0", v.Major, v.Minor, v.Patch)
		}
	})

	t.Run("IncrementMinor", func(t *testing.T) {
		v := &Version{Major: 1, Minor: 2, Patch: 3}
		v.IncrementMinor()

		if v.Major != 1 || v.Minor != 3 || v.Patch != 0 {
			t.Errorf("IncrementMinor failed: got %d.%d.%d, want 1.3.0", v.Major, v.Minor, v.Patch)
		}
	})

	t.Run("IncrementPatch", func(t *testing.T) {
		v := &Version{Major: 1, Minor: 2, Patch: 3}
		v.IncrementPatch()

		if v.Major != 1 || v.Minor != 2 || v.Patch != 4 {
			t.Errorf("IncrementPatch failed: got %d.%d.%d, want 1.2.4", v.Major, v.Minor, v.Patch)
		}
	})
}

func TestAssemblyVersions(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3}

	if v.AssemblySemVer() != "1.2.3.0" {
		t.Errorf("AssemblySemVer failed: got %s, want 1.2.3.0", v.AssemblySemVer())
	}

	if v.AssemblySemFileVer() != "1.2.3.0" {
		t.Errorf("AssemblySemFileVer failed: got %s, want 1.2.3.0", v.AssemblySemFileVer())
	}

	if v.MajorMinorPatch() != "1.2.3" {
		t.Errorf("MajorMinorPatch failed: got %s, want 1.2.3", v.MajorMinorPatch())
	}
}

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "Simple branch name",
			branch:   "feature",
			expected: "feature",
		},
		{
			name:     "Branch with slash",
			branch:   "feature/user-auth",
			expected: "feature-user-auth",
		},
		{
			name:     "Branch with special characters",
			branch:   "feature/user@auth#test",
			expected: "feature-user-auth-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeBranchName(tt.branch)
			if result != tt.expected {
				t.Errorf("SanitizeBranchName failed: got %s, want %s", result, tt.expected)
			}
		})
	}
}
