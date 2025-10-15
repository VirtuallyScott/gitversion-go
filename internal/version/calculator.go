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
	repo            *git.Repository
	config          *config.Config
	strategyManager *StrategyManager
}

func NewCalculator(repo *git.Repository, cfg *config.Config) *Calculator {
	return &Calculator{
		repo:            repo,
		config:          cfg,
		strategyManager: NewStrategyManager(repo, cfg),
	}
}

func (c *Calculator) CalculateVersion(branch string, workflow WorkflowType, forceIncrement string, nextVersion string) (*semver.Version, error) {
	// Get current branch if not provided
	if branch == "" {
		currentBranch, err := c.repo.GetCurrentBranch()
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		branch = currentBranch
	}

	// Get current commit
	currentCommit, err := c.repo.GetSHA()
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit: %w", err)
	}

	// Get branch configuration
	branchConfig := c.config.GetBranchConfiguration(branch)
	if branchConfig == nil {
		// Fall back to default configuration based on branch type
		branchType := c.getBranchType(branch, workflow)
		branchConfig = c.getDefaultBranchConfig(branchType)
	}

	// Use the strategies system for GitTools/GitVersion compatibility
	var strategiesMask VersionStrategies

	// Add configured version strategy if next version is provided
	if nextVersion != "" {
		strategiesMask |= ConfiguredNextVersion
	}

	// Add default strategies
	strategiesMask |= TaggedCommit | MergeMessage | Fallback

	// Create version context for strategies
	ctx := &VersionContext{
		Repository:    c.repo,
		Config:        c.config,
		CurrentBranch: branch,
		CurrentCommit: currentCommit,
		BranchConfig:  branchConfig,
		NextVersion:   nextVersion,
		Strategies:    strategiesMask,
	}

	// Calculate base versions using strategies
	baseVersions, err := c.strategyManager.GetBaseVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get base versions: %w", err)
	}

	// Find the highest base version
	var baseVersion *BaseVersion
	for _, bv := range baseVersions {
		if baseVersion == nil || bv.SemanticVersion.GreaterThan(baseVersion.SemanticVersion) {
			baseVersion = bv
		}
	}

	if baseVersion == nil {
		// Fallback to 0.0.0 if no base version found
		version := &semver.Version{Major: 0, Minor: 0, Patch: 0}
		baseVersion = &BaseVersion{
			Source:            "fallback",
			SemanticVersion:   version,
			ShouldIncrement:   true,
			BaseVersionSource: "fallback",
		}
	}

	// Apply increments based on configuration
	version := baseVersion.SemanticVersion.Copy()

	// Handle force increment
	if forceIncrement != "" {
		switch forceIncrement {
		case "major":
			version.IncrementMajor()
		case "minor":
			version.IncrementMinor()
		case "patch":
			version.IncrementPatch()
		}
	} else if branchConfig.PreventIncrement == nil || (!branchConfig.PreventIncrement.OfMergedBranch && !branchConfig.PreventIncrement.WhenCurrentCommitTagged) {
		// Apply default increment if not prevented
		increment := branchConfig.Increment
		switch increment {
		case config.IncrementMajor:
			version.IncrementMajor()
		case config.IncrementMinor:
			version.IncrementMinor()
		case config.IncrementPatch, "":
			version.IncrementPatch()
		}
	}

	// Apply branch-specific versioning (prerelease, build metadata)
	branchType := c.getBranchType(branch, workflow)
	commitCount, err := c.repo.GetCommitCountSinceTag("")
	if err != nil {
		commitCount = 0
	}

	sha, err := c.repo.GetShortSHA()
	if err != nil {
		sha = "unknown"
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

func (c *Calculator) getDefaultBranchConfig(branchType BranchType) *config.BranchConfiguration {
	switch branchType {
	case Main:
		return &config.BranchConfiguration{
			Increment: config.IncrementPatch,
			PreventIncrement: &config.PreventIncrementConfiguration{
				OfMergedBranch:          false,
				WhenCurrentCommitTagged: false,
				WhenBranchMerged:        false,
			},
			Regex:                 "^(master|main)$",
			SourceBranches:        []string{},
			IsMainBranch:          true,
			PreReleaseWeight:      55000,
			TracksReleaseBranches: false,
		}
	case Develop:
		return &config.BranchConfiguration{
			Increment: config.IncrementMinor,
			PreventIncrement: &config.PreventIncrementConfiguration{
				OfMergedBranch:          false,
				WhenCurrentCommitTagged: false,
				WhenBranchMerged:        false,
			},
			Regex:                 "^develop$",
			SourceBranches:        []string{},
			IsMainBranch:          false,
			PreReleaseWeight:      0,
			TracksReleaseBranches: true,
		}
	case Feature:
		return &config.BranchConfiguration{
			Increment: config.IncrementPatch,
			PreventIncrement: &config.PreventIncrementConfiguration{
				OfMergedBranch:          false,
				WhenCurrentCommitTagged: false,
				WhenBranchMerged:        false,
			},
			Regex:                 "^feature/.+",
			SourceBranches:        []string{"develop", "main", "master"},
			IsMainBranch:          false,
			PreReleaseWeight:      30000,
			TracksReleaseBranches: false,
		}
	case Release:
		return &config.BranchConfiguration{
			Increment: config.IncrementPatch,
			PreventIncrement: &config.PreventIncrementConfiguration{
				OfMergedBranch:          false,
				WhenCurrentCommitTagged: true,
				WhenBranchMerged:        false,
			},
			Regex:                 "^release/.+",
			SourceBranches:        []string{"develop"},
			IsMainBranch:          false,
			PreReleaseWeight:      25000,
			TracksReleaseBranches: false,
		}
	case Hotfix:
		return &config.BranchConfiguration{
			Increment: config.IncrementPatch,
			PreventIncrement: &config.PreventIncrementConfiguration{
				OfMergedBranch:          false,
				WhenCurrentCommitTagged: false,
				WhenBranchMerged:        false,
			},
			Regex:                 "^hotfix/.+",
			SourceBranches:        []string{"main", "master"},
			IsMainBranch:          false,
			PreReleaseWeight:      40000,
			TracksReleaseBranches: false,
		}
	default:
		return &config.BranchConfiguration{
			Increment: config.IncrementPatch,
			PreventIncrement: &config.PreventIncrementConfiguration{
				OfMergedBranch:          false,
				WhenCurrentCommitTagged: false,
				WhenBranchMerged:        false,
			},
			Regex:                 ".*",
			SourceBranches:        []string{},
			IsMainBranch:          false,
			PreReleaseWeight:      30000,
			TracksReleaseBranches: false,
		}
	}
}
