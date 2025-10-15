package git

import (
	"regexp"
	"testing"
)

var (
	semverMajorPattern          = regexp.MustCompile(`(?i)\+semver:\s*(breaking|major)`)
	semverMinorPattern          = regexp.MustCompile(`(?i)\+semver:\s*(feature|minor)`)
	breakingChangePattern       = regexp.MustCompile(`(?i)BREAKING\s*CHANGE`)
	conventionalBreakingPattern = regexp.MustCompile(`(?i)^feat(\(.+\))?!:`)
	conventionalFeaturePattern  = regexp.MustCompile(`(?i)^feat(\(.+\))?:`)
)

func TestIncrementType(t *testing.T) {
	tests := []struct {
		name              string
		commitMessages    []string
		expectedIncrement IncrementType
	}{
		{
			name: "Patch increment for fix",
			commitMessages: []string{
				"fix: resolve login issue",
			},
			expectedIncrement: IncrementPatch,
		},
		{
			name: "Minor increment for feature",
			commitMessages: []string{
				"feat: add user authentication",
			},
			expectedIncrement: IncrementMinor,
		},
		{
			name: "Major increment for breaking change",
			commitMessages: []string{
				"feat!: redesign API",
			},
			expectedIncrement: IncrementMajor,
		},
		{
			name: "Major increment for semver tag",
			commitMessages: []string{
				"fix: critical bug +semver: major",
			},
			expectedIncrement: IncrementMajor,
		},
		{
			name: "Minor increment for semver tag",
			commitMessages: []string{
				"update: improve performance +semver: minor",
			},
			expectedIncrement: IncrementMinor,
		},
		{
			name: "Major increment for BREAKING CHANGE",
			commitMessages: []string{
				"feat: new feature",
				"",
				"BREAKING CHANGE: API has changed",
			},
			expectedIncrement: IncrementMajor,
		},
		{
			name: "Major takes precedence",
			commitMessages: []string{
				"feat: add feature",
				"fix: bug fix +semver: major",
				"feat: another feature",
			},
			expectedIncrement: IncrementMajor,
		},
		{
			name: "Minor takes precedence over patch",
			commitMessages: []string{
				"fix: bug fix",
				"feat: new feature",
				"fix: another bug fix",
			},
			expectedIncrement: IncrementMinor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &Repository{}
			increment := repo.analyzeCommitMessages(tt.commitMessages)

			if increment != tt.expectedIncrement {
				t.Errorf("Expected %s, got %s", tt.expectedIncrement, increment)
			}
		})
	}
}

func (r *Repository) analyzeCommitMessages(messages []string) IncrementType {
	increment := IncrementPatch

	for _, message := range messages {
		detected := r.detectIncrementFromMessage(message)
		if detected == IncrementMajor {
			return IncrementMajor
		}
		if detected == IncrementMinor && increment != IncrementMajor {
			increment = IncrementMinor
		}
	}

	return increment
}

func (r *Repository) detectIncrementFromMessage(message string) IncrementType {
	if semverMajorPattern.MatchString(message) ||
		breakingChangePattern.MatchString(message) ||
		conventionalBreakingPattern.MatchString(message) {
		return IncrementMajor
	}

	if semverMinorPattern.MatchString(message) ||
		conventionalFeaturePattern.MatchString(message) {
		return IncrementMinor
	}

	return IncrementPatch
}
