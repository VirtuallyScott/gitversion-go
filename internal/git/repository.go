package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Commit represents a git commit
type Commit struct {
	SHA     string
	Message string
	Date    string
}

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

func (r *Repository) GetTagsOnCurrentBranch() ([]string, error) {
	cmd := exec.Command("git", "tag", "--merged", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, nil
	}

	var tags []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		tag := strings.TrimSpace(scanner.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func (r *Repository) GetCommitSHAForTag(tag string) (string, error) {
	cmd := exec.Command("git", "rev-list", "-n", "1", tag)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	var branches []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.Contains(line, "->") {
			// Remove "origin/" prefix and get just branch name
			branch := strings.TrimPrefix(line, "origin/")
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

func (r *Repository) GetMergeBase(branch1, branch2 string) (string, error) {
	cmd := exec.Command("git", "merge-base", branch1, branch2)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetCommitHistory(limit int) ([]*Commit, error) {
	cmd := exec.Command("git", "log", "--format=%H|%s|%ci", fmt.Sprintf("-%d", limit))
	output, err := cmd.Output()
	if err != nil {
		return []*Commit{}, err
	}

	var commits []*Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			parts := strings.SplitN(line, "|", 3)
			if len(parts) == 3 {
				commits = append(commits, &Commit{
					SHA:     parts[0],
					Message: parts[1],
					Date:    parts[2],
				})
			}
		}
	}

	return commits, nil
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
