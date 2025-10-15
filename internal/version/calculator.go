package version

import (
	"fmt"
	"strings"

	"github.com/VirtuallyScott/gitversion-go/internal/git"
	"github.com/VirtuallyScott/gitversion-go/pkg/config"
	"github.com/VirtuallyScott/gitversion-go/pkg/semver"
)

type WorkflowType string

const (
	GitFlow    WorkflowType = "gitflow"
	GitHubFlow WorkflowType = "githubflow"
	Trunk      WorkflowType = "trunk"
)

type BranchType string

const (
	Main    BranchType = "main"
	Develop BranchType = "develop"
	Feature BranchType = "feature"
	Release BranchType = "release"
	Hotfix  BranchType = "hotfix"
	Support BranchType = "support"
	Unknown BranchType = "unknown"
)

type Calculator struct {
	repo   *git.Repository
	config *config.Config
}

func NewCalculator(repo *git.Repository, cfg *config.Config) *Calculator {
	return &Calculator{
		repo:   repo,
		config: cfg,
	}
}

func (c *Calculator) CalculateVersion(branch string, workflow WorkflowType, forceIncrement string, nextVersion string) (*semver.Version, error) {
	latestTag, err := c.repo.GetLatestTag()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %w", err)
	}

	var version *semver.Version
	if latestTag != "" {
		version, err = semver.Parse(latestTag)
		if err != nil {
			version = &semver.Version{Major: 0, Minor: 0, Patch: 0}
		}
	} else {
		version = &semver.Version{Major: 0, Minor: 0, Patch: 0}
	}

	if nextVersion != "" {
		version, err = semver.Parse(nextVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid next-version format: %s", nextVersion)
		}
	}

	branchType := c.getBranchType(branch, workflow)

	var increment git.IncrementType
	if forceIncrement != "" {
		switch forceIncrement {
		case "major":
			increment = git.IncrementMajor
		case "minor":
			increment = git.IncrementMinor
		case "patch":
			increment = git.IncrementPatch
		default:
			increment = git.IncrementPatch
		}
	} else {
		increment, err = c.repo.DetectVersionIncrement(latestTag)
		if err != nil {
			return nil, fmt.Errorf("failed to detect version increment: %w", err)
		}
	}

	switch increment {
	case git.IncrementMajor:
		version.IncrementMajor()
	case git.IncrementMinor:
		version.IncrementMinor()
	case git.IncrementPatch:
		version.IncrementPatch()
	}

	commitCount, err := c.repo.GetCommitCountSinceTag(latestTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit count: %w", err)
	}

	sha, err := c.repo.GetShortSHA()
	if err != nil {
		return nil, fmt.Errorf("failed to get SHA: %w", err)
	}

	c.applyBranchSpecificVersioning(version, branch, branchType, commitCount, sha)

	return version, nil
}

func (c *Calculator) getBranchType(branch string, workflow WorkflowType) BranchType {
	switch workflow {
	case GitFlow:
		switch {
		case branch == "main" || branch == "master":
			return Main
		case branch == "develop":
			return Develop
		case strings.HasPrefix(branch, "feature/"):
			return Feature
		case strings.HasPrefix(branch, "release/"):
			return Release
		case strings.HasPrefix(branch, "hotfix/"):
			return Hotfix
		case strings.HasPrefix(branch, "support/"):
			return Support
		default:
			return Unknown
		}
	case GitHubFlow:
		if branch == "main" || branch == "master" {
			return Main
		}
		return Feature
	case Trunk:
		return Main
	default:
		return Unknown
	}
}

func (c *Calculator) applyBranchSpecificVersioning(version *semver.Version, branch string, branchType BranchType, commitCount int, sha string) {
	switch branchType {
	case Main:
		version.Build = fmt.Sprintf("%d+%s", commitCount, sha)
	case Develop:
		if commitCount > 0 {
			version.PreRelease = fmt.Sprintf("alpha.%d", commitCount)
		}
		version.Build = fmt.Sprintf("%d+%s", commitCount, sha)
	case Feature:
		if commitCount > 0 {
			featureName := c.extractFeatureName(branch)
			version.PreRelease = fmt.Sprintf("%s.%d", featureName, commitCount)
		}
		version.Build = fmt.Sprintf("%d+%s", commitCount, sha)
	case Release:
		if commitCount > 0 {
			// Extract prerelease tag from branch name (e.g., release/0.0.2-alpha -> alpha)
			releaseName := c.extractReleaseName(branch)
			if releaseName != "" {
				version.PreRelease = fmt.Sprintf("%s.%d", releaseName, commitCount)
			} else {
				version.PreRelease = fmt.Sprintf("beta.%d", commitCount)
			}
		}
		version.Build = fmt.Sprintf("%d+%s", commitCount, sha)
	case Hotfix:
		if commitCount > 0 {
			version.PreRelease = fmt.Sprintf("hotfix.%d", commitCount)
		}
		version.Build = fmt.Sprintf("%d+%s", commitCount, sha)
	default:
		if commitCount > 0 {
			safeBranch := semver.SanitizeBranchName(branch)
			version.PreRelease = fmt.Sprintf("%s.%d", safeBranch, commitCount)
		}
		version.Build = fmt.Sprintf("%d+%s", commitCount, sha)
	}
}

func (c *Calculator) extractFeatureName(branch string) string {
	parts := strings.Split(branch, "/")
	if len(parts) > 1 {
		return semver.SanitizeBranchName(parts[len(parts)-1])
	}
	return semver.SanitizeBranchName(branch)
}

func (c *Calculator) extractReleaseName(branch string) string {
	parts := strings.Split(branch, "/")
	if len(parts) > 1 {
		versionPart := parts[len(parts)-1]
		// Look for prerelease tag after a dash (e.g., "0.0.2-alpha" -> "alpha")  
		if dashIndex := strings.LastIndex(versionPart, "-"); dashIndex != -1 && dashIndex < len(versionPart)-1 {
			prerelease := versionPart[dashIndex+1:]
			return semver.SanitizeBranchName(prerelease)
		}
	}
	return ""
}
