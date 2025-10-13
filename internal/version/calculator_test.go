package version

import (
	"testing"

	"github.com/VirtuallyScott/battle-tested-devops/gitversion-go/pkg/semver"
)

func TestGetBranchType(t *testing.T) {
	calculator := &Calculator{}
	
	tests := []struct {
		name     string
		branch   string
		workflow WorkflowType
		expected BranchType
	}{
		// GitFlow tests
		{
			name:     "GitFlow main branch",
			branch:   "main",
			workflow: GitFlow,
			expected: Main,
		},
		{
			name:     "GitFlow master branch",
			branch:   "master",
			workflow: GitFlow,
			expected: Main,
		},
		{
			name:     "GitFlow develop branch",
			branch:   "develop",
			workflow: GitFlow,
			expected: Develop,
		},
		{
			name:     "GitFlow feature branch",
			branch:   "feature/user-auth",
			workflow: GitFlow,
			expected: Feature,
		},
		{
			name:     "GitFlow release branch",
			branch:   "release/1.2.0",
			workflow: GitFlow,
			expected: Release,
		},
		{
			name:     "GitFlow hotfix branch",
			branch:   "hotfix/critical-fix",
			workflow: GitFlow,
			expected: Hotfix,
		},
		{
			name:     "GitFlow support branch",
			branch:   "support/v1.x",
			workflow: GitFlow,
			expected: Support,
		},
		{
			name:     "GitFlow unknown branch",
			branch:   "random-branch",
			workflow: GitFlow,
			expected: Unknown,
		},
		
		// GitHubFlow tests
		{
			name:     "GitHubFlow main branch",
			branch:   "main",
			workflow: GitHubFlow,
			expected: Main,
		},
		{
			name:     "GitHubFlow feature branch",
			branch:   "feature/user-auth",
			workflow: GitHubFlow,
			expected: Feature,
		},
		{
			name:     "GitHubFlow any other branch",
			branch:   "some-feature",
			workflow: GitHubFlow,
			expected: Feature,
		},
		
		// Trunk tests
		{
			name:     "Trunk any branch",
			branch:   "any-branch",
			workflow: Trunk,
			expected: Main,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.getBranchType(tt.branch, tt.workflow)
			if result != tt.expected {
				t.Errorf("getBranchType(%s, %s) = %s, want %s", tt.branch, tt.workflow, result, tt.expected)
			}
		})
	}
}

func TestExtractFeatureName(t *testing.T) {
	calculator := &Calculator{}
	
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "Simple feature branch",
			branch:   "feature/user-auth",
			expected: "user-auth",
		},
		{
			name:     "Complex feature branch",
			branch:   "feature/complex/user-auth",
			expected: "user-auth",
		},
		{
			name:     "Feature branch with special characters",
			branch:   "feature/user@auth#test",
			expected: "user-auth-test",
		},
		{
			name:     "Simple branch name",
			branch:   "simple-branch",
			expected: "simple-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.extractFeatureName(tt.branch)
			if result != tt.expected {
				t.Errorf("extractFeatureName(%s) = %s, want %s", tt.branch, result, tt.expected)
			}
		})
	}
}

func TestApplyBranchSpecificVersioning(t *testing.T) {
	calculator := &Calculator{}
	
	tests := []struct {
		name              string
		branchType        BranchType
		branch            string
		commitCount       int
		sha               string
		expectedPreRelease string
		expectedBuild     string
	}{
		{
			name:              "Main branch",
			branchType:        Main,
			branch:            "main",
			commitCount:       5,
			sha:               "abc123",
			expectedPreRelease: "",
			expectedBuild:     "5+abc123",
		},
		{
			name:              "Develop branch with commits",
			branchType:        Develop,
			branch:            "develop",
			commitCount:       10,
			sha:               "def456",
			expectedPreRelease: "alpha.10",
			expectedBuild:     "10+def456",
		},
		{
			name:              "Develop branch without commits",
			branchType:        Develop,
			branch:            "develop",
			commitCount:       0,
			sha:               "ghi789",
			expectedPreRelease: "",
			expectedBuild:     "0+ghi789",
		},
		{
			name:              "Feature branch with commits",
			branchType:        Feature,
			branch:            "feature/user-auth",
			commitCount:       3,
			sha:               "jkl012",
			expectedPreRelease: "user-auth.3",
			expectedBuild:     "3+jkl012",
		},
		{
			name:              "Release branch with commits",
			branchType:        Release,
			branch:            "release/1.2.0",
			commitCount:       2,
			sha:               "mno345",
			expectedPreRelease: "beta.2",
			expectedBuild:     "2+mno345",
		},
		{
			name:              "Hotfix branch with commits",
			branchType:        Hotfix,
			branch:            "hotfix/critical",
			commitCount:       1,
			sha:               "pqr678",
			expectedPreRelease: "hotfix.1",
			expectedBuild:     "1+pqr678",
		},
		{
			name:              "Unknown branch with commits",
			branchType:        Unknown,
			branch:            "custom-branch",
			commitCount:       4,
			sha:               "stu901",
			expectedPreRelease: "custom-branch.4",
			expectedBuild:     "4+stu901",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := &semver.Version{Major: 1, Minor: 2, Patch: 3}
			calculator.applyBranchSpecificVersioning(version, tt.branch, tt.branchType, tt.commitCount, tt.sha)
			
			if version.PreRelease != tt.expectedPreRelease {
				t.Errorf("PreRelease = %s, want %s", version.PreRelease, tt.expectedPreRelease)
			}
			if version.Build != tt.expectedBuild {
				t.Errorf("Build = %s, want %s", version.Build, tt.expectedBuild)
			}
		})
	}
}