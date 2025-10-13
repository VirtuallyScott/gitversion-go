package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) IsRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "HEAD", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetLatestTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetCommitCountSinceTag(tag string) (int, error) {
	var cmd *exec.Cmd
	if tag != "" {
		cmd = exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s..HEAD", tag))
	} else {
		cmd = exec.Command("git", "rev-list", "--count", "HEAD")
	}
	
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}
	
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, nil
	}
	
	return count, nil
}

func (r *Repository) GetCommitsSinceTag(tag string) ([]string, error) {
	var cmd *exec.Cmd
	if tag != "" {
		cmd = exec.Command("git", "log", "--oneline", fmt.Sprintf("%s..HEAD", tag))
	} else {
		cmd = exec.Command("git", "log", "--oneline", "HEAD")
	}
	
	output, err := cmd.Output()
	if err != nil {
		return []string{}, nil
	}
	
	var commits []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commits = append(commits, line)
		}
	}
	
	return commits, nil
}

func (r *Repository) GetShortSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetCommitDate() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ci", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}
	return strings.TrimSpace(string(output)), nil
}

type IncrementType string

const (
	IncrementPatch IncrementType = "patch"
	IncrementMinor IncrementType = "minor"
	IncrementMajor IncrementType = "major"
)

func (r *Repository) DetectVersionIncrement(tag string) (IncrementType, error) {
	commits, err := r.GetCommitsSinceTag(tag)
	if err != nil {
		return IncrementPatch, err
	}
	
	increment := IncrementPatch
	
	semverMajorPattern := regexp.MustCompile(`(?i)\+semver:\s*(breaking|major)`)
	semverMinorPattern := regexp.MustCompile(`(?i)\+semver:\s*(feature|minor)`)
	breakingChangePattern := regexp.MustCompile(`(?i)BREAKING\s*CHANGE`)
	conventionalBreakingPattern := regexp.MustCompile(`(?i)^feat(\(.+\))?!:`)
	conventionalFeaturePattern := regexp.MustCompile(`(?i)^feat(\(.+\))?:`)
	
	for _, commit := range commits {
		if semverMajorPattern.MatchString(commit) || 
		   breakingChangePattern.MatchString(commit) || 
		   conventionalBreakingPattern.MatchString(commit) {
			return IncrementMajor, nil
		}
		
		if semverMinorPattern.MatchString(commit) || 
		   conventionalFeaturePattern.MatchString(commit) {
			if increment != IncrementMajor {
				increment = IncrementMinor
			}
		}
	}
	
	return increment, nil
}