package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// IncrementStrategy represents how version increments are applied
type IncrementStrategy string

const (
	IncrementNone    IncrementStrategy = "None"
	IncrementPatch   IncrementStrategy = "Patch"
	IncrementMinor   IncrementStrategy = "Minor"
	IncrementMajor   IncrementStrategy = "Major"
	IncrementInherit IncrementStrategy = "Inherit"
)

// DeploymentMode represents the deployment mode for a branch
type DeploymentMode string

const (
	DeploymentManual             DeploymentMode = "ManualDeployment"
	DeploymentContinuousDelivery DeploymentMode = "ContinuousDelivery"
	DeploymentContinuous         DeploymentMode = "ContinuousDeployment"
)

// PreventIncrementConfiguration defines when to prevent version increments
type PreventIncrementConfiguration struct {
	OfMergedBranch          bool `json:"of-merged-branch" yaml:"of-merged-branch"`
	WhenCurrentCommitTagged bool `json:"when-current-commit-tagged" yaml:"when-current-commit-tagged"`
	WhenBranchMerged        bool `json:"when-branch-merged" yaml:"when-branch-merged"`
}

// BranchConfiguration represents configuration for a specific branch type
type BranchConfiguration struct {
	Mode                  DeploymentMode                 `json:"mode" yaml:"mode"`
	Label                 string                         `json:"label" yaml:"label"`
	Increment             IncrementStrategy              `json:"increment" yaml:"increment"`
	PreventIncrement      *PreventIncrementConfiguration `json:"prevent-increment" yaml:"prevent-increment"`
	TrackMergeTarget      bool                           `json:"track-merge-target" yaml:"track-merge-target"`
	TrackMergeMessage     bool                           `json:"track-merge-message" yaml:"track-merge-message"`
	Regex                 string                         `json:"regex" yaml:"regex"`
	SourceBranches        []string                       `json:"source-branches" yaml:"source-branches"`
	IsSourceBranchFor     []string                       `json:"is-source-branch-for" yaml:"is-source-branch-for"`
	TracksReleaseBranches bool                           `json:"tracks-release-branches" yaml:"tracks-release-branches"`
	IsReleaseBranch       bool                           `json:"is-release-branch" yaml:"is-release-branch"`
	IsMainBranch          bool                           `json:"is-main-branch" yaml:"is-main-branch"`
	PreReleaseWeight      int                            `json:"pre-release-weight" yaml:"pre-release-weight"`
}

// Legacy BranchConfig for backward compatibility
type BranchConfig struct {
	Increment string `json:"increment" yaml:"increment"`
	Tag       string `json:"tag" yaml:"tag"`
	Regex     string `json:"regex" yaml:"regex"`
}

type CommitMessageConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	IncrementMode string `json:"increment-mode" yaml:"increment-mode"`
}

type Config struct {
	NextVersion             string                          `json:"next-version" yaml:"next-version"`
	Mode                    DeploymentMode                  `json:"mode" yaml:"mode"`
	Increment               IncrementStrategy               `json:"increment" yaml:"increment"`
	TagPrefix               string                          `json:"tag-prefix" yaml:"tag-prefix"`
	MajorVersionBumpMessage string                          `json:"major-version-bump-message" yaml:"major-version-bump-message"`
	MinorVersionBumpMessage string                          `json:"minor-version-bump-message" yaml:"minor-version-bump-message"`
	PatchVersionBumpMessage string                          `json:"patch-version-bump-message" yaml:"patch-version-bump-message"`
	NoBumpMessage           string                          `json:"no-bump-message" yaml:"no-bump-message"`
	TagPreReleaseWeight     int                             `json:"tag-pre-release-weight" yaml:"tag-pre-release-weight"`
	CommitDateFormat        string                          `json:"commit-date-format" yaml:"commit-date-format"`
	MergeMessageFormats     map[string]interface{}          `json:"merge-message-formats" yaml:"merge-message-formats"`
	UpdateBuildNumber       bool                            `json:"update-build-number" yaml:"update-build-number"`
	SemanticVersionFormat   string                          `json:"semantic-version-format" yaml:"semantic-version-format"`
	Strategies              []string                        `json:"strategies" yaml:"strategies"`
	Branches                map[string]*BranchConfiguration `json:"branches" yaml:"branches"`
	Ignore                  map[string][]string             `json:"ignore" yaml:"ignore"`
	CommitMessageIncrement  CommitMessageConfig             `json:"commit-message-incrementing" yaml:"commit-message-incrementing"`

	// Legacy fields for backward compatibility
	LegacyBranches map[string]BranchConfig `json:"-" yaml:"-"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return getDefaultConfig(), nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported configuration file format: %s", ext)
	}

	// Set defaults for empty fields
	if config.Mode == "" {
		config.Mode = DeploymentContinuousDelivery
	}
	if config.Increment == "" {
		config.Increment = IncrementInherit
	}
	if config.TagPrefix == "" {
		config.TagPrefix = "[vV]"
	}
	if config.TagPreReleaseWeight == 0 {
		config.TagPreReleaseWeight = 60000
	}
	if config.CommitDateFormat == "" {
		config.CommitDateFormat = "yyyy-MM-dd"
	}
	if config.SemanticVersionFormat == "" {
		config.SemanticVersionFormat = "Strict"
	}
	if len(config.Strategies) == 0 {
		config.Strategies = []string{
			"Fallback",
			"ConfiguredNextVersion",
			"MergeMessage",
			"TaggedCommit",
			"TrackReleaseBranches",
			"VersionInBranchName",
		}
	}

	// Initialize branch configurations if not present
	if config.Branches == nil {
		config.Branches = getDefaultBranchConfigurations()
	}

	return config, nil
}

func getDefaultConfig() *Config {
	return &Config{
		NextVersion:             "1.0.0",
		Mode:                    DeploymentContinuousDelivery,
		Increment:               IncrementInherit,
		TagPrefix:               "[vV]",
		MajorVersionBumpMessage: `\+semver:\s?(breaking|major)`,
		MinorVersionBumpMessage: `\+semver:\s?(feature|minor)`,
		PatchVersionBumpMessage: `\+semver:\s?(fix|patch)`,
		NoBumpMessage:           `\+semver:\s?(none|skip)`,
		TagPreReleaseWeight:     60000,
		CommitDateFormat:        "yyyy-MM-dd",
		UpdateBuildNumber:       true,
		SemanticVersionFormat:   "Strict",
		Strategies: []string{
			"Fallback",
			"ConfiguredNextVersion",
			"MergeMessage",
			"TaggedCommit",
			"TrackReleaseBranches",
			"VersionInBranchName",
		},
		Branches: getDefaultBranchConfigurations(),
		Ignore: map[string][]string{
			"sha": {},
		},
		MergeMessageFormats: map[string]interface{}{},
		CommitMessageIncrement: CommitMessageConfig{
			Enabled:       true,
			IncrementMode: "Enabled",
		},
	}
}

func getDefaultBranchConfigurations() map[string]*BranchConfiguration {
	return map[string]*BranchConfiguration{
		"main": {
			Mode:                  DeploymentManual,
			Label:                 "",
			Increment:             IncrementPatch,
			PreventIncrement:      &PreventIncrementConfiguration{WhenCurrentCommitTagged: false},
			TrackMergeTarget:      false,
			TrackMergeMessage:     true,
			Regex:                 "^(master|main)$",
			SourceBranches:        []string{},
			IsSourceBranchFor:     []string{},
			TracksReleaseBranches: false,
			IsReleaseBranch:       false,
			IsMainBranch:          true,
			PreReleaseWeight:      55000,
		},
		"develop": {
			Mode:                  DeploymentContinuousDelivery,
			Label:                 "alpha",
			Increment:             IncrementMinor,
			PreventIncrement:      &PreventIncrementConfiguration{WhenCurrentCommitTagged: false},
			TrackMergeTarget:      true,
			TrackMergeMessage:     true,
			Regex:                 "^dev(elop)?(ment)?$",
			SourceBranches:        []string{"main"},
			IsSourceBranchFor:     []string{},
			TracksReleaseBranches: true,
			IsReleaseBranch:       false,
			IsMainBranch:          false,
			PreReleaseWeight:      0,
		},
		"release": {
			Mode:      DeploymentManual,
			Label:     "beta",
			Increment: IncrementNone,
			PreventIncrement: &PreventIncrementConfiguration{
				OfMergedBranch:          true,
				WhenCurrentCommitTagged: false,
			},
			TrackMergeTarget:      false,
			TrackMergeMessage:     true,
			Regex:                 `^releases?[\/-](?<BranchName>.+)`,
			SourceBranches:        []string{"main", "support"},
			IsSourceBranchFor:     []string{},
			TracksReleaseBranches: false,
			IsReleaseBranch:       true,
			IsMainBranch:          false,
			PreReleaseWeight:      30000,
		},
		"feature": {
			Mode:                  DeploymentManual,
			Label:                 "{BranchName}",
			Increment:             IncrementInherit,
			PreventIncrement:      &PreventIncrementConfiguration{WhenCurrentCommitTagged: false},
			TrackMergeTarget:      false,
			TrackMergeMessage:     true,
			Regex:                 `^features?[\/-](?<BranchName>.+)`,
			SourceBranches:        []string{"develop", "main", "release", "support", "hotfix"},
			IsSourceBranchFor:     []string{},
			TracksReleaseBranches: false,
			IsReleaseBranch:       false,
			IsMainBranch:          false,
			PreReleaseWeight:      30000,
		},
		"pull-request": {
			Mode:      DeploymentContinuousDelivery,
			Label:     "PullRequest{Number}",
			Increment: IncrementInherit,
			PreventIncrement: &PreventIncrementConfiguration{
				OfMergedBranch:          true,
				WhenCurrentCommitTagged: false,
			},
			TrackMergeMessage: true,
			Regex:             `^(pull-requests|pull|pr)[\/-](?<Number>\d*)`,
			SourceBranches:    []string{"develop", "main", "release", "feature", "support", "hotfix"},
			IsSourceBranchFor: []string{},
			PreReleaseWeight:  30000,
		},
		"hotfix": {
			Mode:                  DeploymentManual,
			Label:                 "beta",
			Increment:             IncrementInherit,
			PreventIncrement:      &PreventIncrementConfiguration{WhenCurrentCommitTagged: false},
			Regex:                 `^hotfix(es)?[\/-](?<BranchName>.+)`,
			SourceBranches:        []string{"main", "support"},
			IsSourceBranchFor:     []string{},
			TracksReleaseBranches: false,
			IsReleaseBranch:       true,
			IsMainBranch:          false,
			PreReleaseWeight:      30000,
		},
		"support": {
			Label:             "",
			Increment:         IncrementPatch,
			PreventIncrement:  &PreventIncrementConfiguration{WhenCurrentCommitTagged: false},
			TrackMergeTarget:  false,
			Regex:             `^support[\/-](?<BranchName>.+)`,
			SourceBranches:    []string{"main"},
			IsSourceBranchFor: []string{},
			IsMainBranch:      false,
			PreReleaseWeight:  55000,
		},
	}
}

func (c *Config) GetBranchConfig(branchName string) *BranchConfig {
	// First try legacy branches for backward compatibility
	if c.LegacyBranches != nil {
		if config, exists := c.LegacyBranches[branchName]; exists {
			return &config
		}

		for branchType, config := range c.LegacyBranches {
			if strings.HasPrefix(branchName, branchType+"/") {
				return &config
			}
		}
	}

	return &BranchConfig{
		Increment: "Patch",
		Tag:       "{BranchName}",
		Regex:     "",
	}
}

func (c *Config) GetBranchConfiguration(branchName string) *BranchConfiguration {
	// Try exact match first
	if config, exists := c.Branches[branchName]; exists {
		return config
	}

	// Try regex matching
	for _, config := range c.Branches {
		if config.Regex != "" && matchesRegex(branchName, config.Regex) {
			return config
		}
	}

	// Try prefix matching as fallback
	for branchType, config := range c.Branches {
		if strings.HasPrefix(branchName, branchType+"/") {
			return config
		}
	}

	// Return default configuration
	return &BranchConfiguration{
		Mode:              DeploymentManual,
		Label:             "{BranchName}",
		Increment:         IncrementPatch,
		PreventIncrement:  &PreventIncrementConfiguration{WhenCurrentCommitTagged: false},
		Regex:             "",
		SourceBranches:    []string{},
		IsSourceBranchFor: []string{},
		IsMainBranch:      false,
		PreReleaseWeight:  30000,
	}
}

func matchesRegex(branchName, pattern string) bool {
	// Simple regex matching - in a real implementation you'd use regexp package
	// For now, handle basic cases
	if pattern == "^(master|main)$" {
		return branchName == "master" || branchName == "main"
	}
	if pattern == "^dev(elop)?(ment)?$" {
		return branchName == "dev" || branchName == "develop" || branchName == "development"
	}
	if strings.Contains(pattern, "releases?") {
		return strings.HasPrefix(branchName, "release/") || strings.HasPrefix(branchName, "releases/")
	}
	if strings.Contains(pattern, "features?") {
		return strings.HasPrefix(branchName, "feature/") || strings.HasPrefix(branchName, "features/")
	}
	if strings.Contains(pattern, "hotfix") {
		return strings.HasPrefix(branchName, "hotfix/") || strings.HasPrefix(branchName, "hotfixes/")
	}
	if strings.Contains(pattern, "support") {
		return strings.HasPrefix(branchName, "support/")
	}
	if strings.Contains(pattern, "pull") {
		return strings.HasPrefix(branchName, "pull/") || strings.HasPrefix(branchName, "pr/")
	}
	return false
}
