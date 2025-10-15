package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/VirtuallyScott/gitversion-go/internal/git"
	"github.com/VirtuallyScott/gitversion-go/pkg/config"
	"github.com/VirtuallyScott/gitversion-go/pkg/semver"
)

// VersionStrategies represents the available version calculation strategies
type VersionStrategies int

const (
	// None indicates no strategy
	None VersionStrategies = 0
	// Fallback strategy - always returns 0.0.0
	Fallback VersionStrategies = 1 << iota
	// ConfiguredNextVersion strategy - returns version from config
	ConfiguredNextVersion
	// MergeMessage strategy - extracts version from merge commit messages
	MergeMessage
	// TaggedCommit strategy - extracts version from tags
	TaggedCommit
	// TrackReleaseBranches strategy - tracks versions from release branches
	TrackReleaseBranches
	// VersionInBranchName strategy - extracts version from branch name
	VersionInBranchName
	// Mainline strategy - increments version on every commit for main branches
	Mainline
)

// BaseVersion represents a version source with metadata
type BaseVersion struct {
	SemanticVersion   *semver.Version
	Source            string
	ShouldIncrement   bool
	BaseVersionSource string
}

// VersionStrategy defines the interface for version calculation strategies
type VersionStrategy interface {
	GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error)
	GetName() string
}

// VersionContext provides context for version calculation
type VersionContext struct {
	Repository    *git.Repository
	Config        *config.Config
	CurrentBranch string
	CurrentCommit string
	BranchConfig  *config.BranchConfiguration
	Strategies    VersionStrategies
	NextVersion   string
}

// FallbackStrategy implements the fallback version strategy
type FallbackStrategy struct{}

func (f *FallbackStrategy) GetName() string {
	return "Fallback"
}

func (f *FallbackStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	return []*BaseVersion{
		{
			SemanticVersion:   &semver.Version{Major: 0, Minor: 0, Patch: 0},
			Source:            "Fallback strategy",
			ShouldIncrement:   true,
			BaseVersionSource: "",
		},
	}, nil
}

// ConfiguredNextVersionStrategy implements the configured next version strategy
type ConfiguredNextVersionStrategy struct{}

func (c *ConfiguredNextVersionStrategy) GetName() string {
	return "ConfiguredNextVersion"
}

func (c *ConfiguredNextVersionStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	nextVersion := ctx.NextVersion
	if nextVersion == "" {
		nextVersion = ctx.Config.NextVersion
	}

	if nextVersion == "" {
		return nil, nil
	}

	version, err := semver.Parse(nextVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid next-version: %w", err)
	}

	return []*BaseVersion{
		{
			SemanticVersion:   version,
			Source:            fmt.Sprintf("Configured next version: %s", nextVersion),
			ShouldIncrement:   false,
			BaseVersionSource: "",
		},
	}, nil
}

// TaggedCommitStrategy implements the tagged commit strategy
type TaggedCommitStrategy struct{}

func (t *TaggedCommitStrategy) GetName() string {
	return "TaggedCommit"
}

func (t *TaggedCommitStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	tags, err := ctx.Repository.GetTagsOnCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var baseVersions []*BaseVersion
	for _, tag := range tags {
		version, err := semver.Parse(tag)
		if err != nil {
			continue // Skip invalid semantic version tags
		}

		sha, err := ctx.Repository.GetCommitSHAForTag(tag)
		if err != nil {
			continue
		}

		baseVersions = append(baseVersions, &BaseVersion{
			SemanticVersion:   version,
			Source:            fmt.Sprintf("Tag '%s'", tag),
			ShouldIncrement:   true,
			BaseVersionSource: sha,
		})
	}

	return baseVersions, nil
}

// MergeMessageStrategy implements the merge message strategy
type MergeMessageStrategy struct{}

func (m *MergeMessageStrategy) GetName() string {
	return "MergeMessage"
}

func (m *MergeMessageStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	if ctx.BranchConfig == nil || !ctx.BranchConfig.TrackMergeMessage {
		return nil, nil
	}

	commits, err := ctx.Repository.GetCommitHistory(50) // Look at recent commits
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	var baseVersions []*BaseVersion
	mergePattern := regexp.MustCompile(`(?i)merge.*?(?:branch\s+)?(?:'([^']+)'|"([^"]+)"|(\S+))`)
	versionPattern := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*))?`)

	for _, commit := range commits {
		matches := mergePattern.FindStringSubmatch(commit.Message)
		if len(matches) == 0 {
			continue
		}

		// Extract branch name from merge message
		branchName := ""
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				branchName = matches[i]
				break
			}
		}

		if branchName == "" {
			continue
		}

		// Look for version in branch name
		versionMatches := versionPattern.FindStringSubmatch(branchName)
		if len(versionMatches) == 0 {
			continue
		}

		major, _ := strconv.Atoi(versionMatches[1])
		minor, _ := strconv.Atoi(versionMatches[2])
		patch, _ := strconv.Atoi(versionMatches[3])
		prerelease := versionMatches[4]

		version := &semver.Version{
			Major: major,
			Minor: minor,
			Patch: patch,
		}
		if prerelease != "" {
			version.PreRelease = prerelease
		}

		shouldIncrement := true
		if ctx.BranchConfig.PreventIncrement != nil && ctx.BranchConfig.PreventIncrement.OfMergedBranch {
			shouldIncrement = false
		}

		baseVersions = append(baseVersions, &BaseVersion{
			SemanticVersion:   version,
			Source:            fmt.Sprintf("Merge message '%s'", commit.Message),
			ShouldIncrement:   shouldIncrement,
			BaseVersionSource: commit.SHA,
		})
	}

	return baseVersions, nil
}

// VersionInBranchNameStrategy implements the version in branch name strategy
type VersionInBranchNameStrategy struct{}

func (v *VersionInBranchNameStrategy) GetName() string {
	return "VersionInBranchName"
}

func (v *VersionInBranchNameStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	versionPattern := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*))?`)
	matches := versionPattern.FindStringSubmatch(ctx.CurrentBranch)

	if len(matches) == 0 {
		return nil, nil
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	prerelease := matches[4]

	version := &semver.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	if prerelease != "" {
		version.PreRelease = prerelease
	}

	return []*BaseVersion{
		{
			SemanticVersion:   version,
			Source:            fmt.Sprintf("Version in branch name '%s'", ctx.CurrentBranch),
			ShouldIncrement:   false,
			BaseVersionSource: ctx.CurrentCommit,
		},
	}, nil
}

// TrackReleaseBranchesStrategy implements the track release branches strategy
type TrackReleaseBranchesStrategy struct{}

func (t *TrackReleaseBranchesStrategy) GetName() string {
	return "TrackReleaseBranches"
}

func (t *TrackReleaseBranchesStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	if ctx.BranchConfig == nil || !ctx.BranchConfig.TracksReleaseBranches {
		return nil, nil
	}

	// Get all branches that match release patterns
	branches, err := ctx.Repository.GetBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	var baseVersions []*BaseVersion
	releasePattern := regexp.MustCompile(`^releases?[/-]`)

	for _, branch := range branches {
		if !releasePattern.MatchString(branch) {
			continue
		}

		// Try to extract version from branch name
		versionStrategy := &VersionInBranchNameStrategy{}
		branchCtx := &VersionContext{
			Repository:    ctx.Repository,
			Config:        ctx.Config,
			CurrentBranch: branch,
			CurrentCommit: ctx.CurrentCommit,
			BranchConfig:  ctx.BranchConfig,
			Strategies:    ctx.Strategies,
		}

		versions, err := versionStrategy.GetBaseVersions(branchCtx)
		if err != nil || len(versions) == 0 {
			continue
		}

		// Find merge base with current branch
		mergeBase, err := ctx.Repository.GetMergeBase(branch, ctx.CurrentBranch)
		if err != nil {
			continue
		}

		for _, version := range versions {
			baseVersions = append(baseVersions, &BaseVersion{
				SemanticVersion:   version.SemanticVersion,
				Source:            fmt.Sprintf("Release branch '%s'", branch),
				ShouldIncrement:   true,
				BaseVersionSource: mergeBase,
			})
		}
	}

	return baseVersions, nil
}

// MainlineStrategy implements the mainline development strategy
type MainlineStrategy struct{}

func (m *MainlineStrategy) GetName() string {
	return "Mainline"
}

func (m *MainlineStrategy) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	if ctx.BranchConfig == nil || !ctx.BranchConfig.IsMainBranch {
		return nil, nil
	}

	// Get the latest tag on this branch
	latestTag, err := ctx.Repository.GetLatestTag()
	if err != nil {
		// If no tags, start from 0.0.0
		return []*BaseVersion{
			{
				SemanticVersion:   &semver.Version{Major: 0, Minor: 0, Patch: 0},
				Source:            "Mainline strategy (no tags)",
				ShouldIncrement:   true,
				BaseVersionSource: "",
			},
		}, nil
	}

	version, err := semver.Parse(latestTag)
	if err != nil {
		// If tag is not a valid semantic version, start from 0.0.0
		return []*BaseVersion{
			{
				SemanticVersion:   &semver.Version{Major: 0, Minor: 0, Patch: 0},
				Source:            "Mainline strategy (invalid tag)",
				ShouldIncrement:   true,
				BaseVersionSource: "",
			},
		}, nil
	}

	tagSHA, err := ctx.Repository.GetCommitSHAForTag(latestTag)
	if err != nil {
		tagSHA = ""
	}

	return []*BaseVersion{
		{
			SemanticVersion:   version,
			Source:            fmt.Sprintf("Mainline strategy from tag '%s'", latestTag),
			ShouldIncrement:   true,
			BaseVersionSource: tagSHA,
		},
	}, nil
}

// StrategyManager manages version calculation strategies
type StrategyManager struct {
	strategies map[VersionStrategies]VersionStrategy
	repo       *git.Repository
	config     *config.Config
}

// NewStrategyManager creates a new strategy manager
func NewStrategyManager(repo *git.Repository, config *config.Config) *StrategyManager {
	return &StrategyManager{
		strategies: map[VersionStrategies]VersionStrategy{
			Fallback:              &FallbackStrategy{},
			ConfiguredNextVersion: &ConfiguredNextVersionStrategy{},
			TaggedCommit:          &TaggedCommitStrategy{},
			MergeMessage:          &MergeMessageStrategy{},
			VersionInBranchName:   &VersionInBranchNameStrategy{},
			TrackReleaseBranches:  &TrackReleaseBranchesStrategy{},
			Mainline:              &MainlineStrategy{},
		},
		repo:   repo,
		config: config,
	}
}

// GetBaseVersions calculates base versions using the specified strategies
func (sm *StrategyManager) GetBaseVersions(ctx *VersionContext) ([]*BaseVersion, error) {
	var allBaseVersions []*BaseVersion

	// Process strategies in order of priority
	strategyOrder := []VersionStrategies{
		ConfiguredNextVersion,
		VersionInBranchName,
		TaggedCommit,
		TrackReleaseBranches,
		MergeMessage,
		Mainline,
		Fallback,
	}

	for _, strategyType := range strategyOrder {
		if ctx.Strategies&strategyType == 0 {
			continue // Strategy not enabled
		}

		strategy, exists := sm.strategies[strategyType]
		if !exists {
			continue
		}

		baseVersions, err := strategy.GetBaseVersions(ctx)
		if err != nil {
			return nil, fmt.Errorf("strategy %s failed: %w", strategy.GetName(), err)
		}

		allBaseVersions = append(allBaseVersions, baseVersions...)
	}

	// If no base versions found, use fallback
	if len(allBaseVersions) == 0 {
		fallback := sm.strategies[Fallback]
		baseVersions, err := fallback.GetBaseVersions(ctx)
		if err != nil {
			return nil, fmt.Errorf("fallback strategy failed: %w", err)
		}
		allBaseVersions = append(allBaseVersions, baseVersions...)
	}

	return allBaseVersions, nil
}

// FindBestBaseVersion selects the best base version from available options
func (sm *StrategyManager) FindBestBaseVersion(baseVersions []*BaseVersion) *BaseVersion {
	if len(baseVersions) == 0 {
		return nil
	}

	// Sort by semantic version (highest first) and prefer non-prerelease versions
	best := baseVersions[0]
	for _, bv := range baseVersions[1:] {
		if bv.SemanticVersion.Compare(best.SemanticVersion) > 0 {
			best = bv
		} else if bv.SemanticVersion.Compare(best.SemanticVersion) == 0 {
			// If versions are equal, prefer the one without prerelease
			if best.SemanticVersion.PreRelease != "" && bv.SemanticVersion.PreRelease == "" {
				best = bv
			}
		}
	}

	return best
}

// ParseVersionStrategies parses strategy strings into the bitwise enum
func ParseVersionStrategies(strategies []string) VersionStrategies {
	var result VersionStrategies

	for _, strategy := range strategies {
		switch strings.ToLower(strings.TrimSpace(strategy)) {
		case "fallback":
			result |= Fallback
		case "configurednextversion":
			result |= ConfiguredNextVersion
		case "mergemessage":
			result |= MergeMessage
		case "taggedcommit":
			result |= TaggedCommit
		case "trackreleasebranches":
			result |= TrackReleaseBranches
		case "versioninbranchname":
			result |= VersionInBranchName
		case "mainline":
			result |= Mainline
		}
	}

	return result
}

// GetDefaultStrategies returns the default set of strategies
func GetDefaultStrategies() VersionStrategies {
	return Fallback | ConfiguredNextVersion | MergeMessage | TaggedCommit | TrackReleaseBranches | VersionInBranchName
}
