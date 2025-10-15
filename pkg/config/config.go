package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	NextVersion            string                         `json:"next-version" yaml:"next-version"`
	Branches               map[string]BranchConfig       `json:"branches" yaml:"branches"`
	Ignore                 map[string][]string            `json:"ignore" yaml:"ignore"`
	MergeMessageFormats    map[string]interface{}         `json:"merge-message-formats" yaml:"merge-message-formats"`
	CommitMessageIncrement CommitMessageConfig            `json:"commit-message-incrementing" yaml:"commit-message-incrementing"`
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

	return config, nil
}

func getDefaultConfig() *Config {
	return &Config{
		NextVersion: "0.0.0",
		Branches: map[string]BranchConfig{
			"main": {
				Increment: "Patch",
				Tag:       "",
				Regex:     "^master$|^main$",
			},
			"develop": {
				Increment: "Minor",
				Tag:       "alpha",
				Regex:     "^develop$",
			},
			"feature": {
				Increment: "Minor",
				Tag:       "{BranchName}",
				Regex:     "^features?[/-]",
			},
			"release": {
				Increment: "None",
				Tag:       "beta",
				Regex:     "^releases?[/-]",
			},
			"hotfix": {
				Increment: "Patch",
				Tag:       "hotfix",
				Regex:     "^hotfix(es)?[/-]",
			},
		},
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

func (c *Config) GetBranchConfig(branchName string) *BranchConfig {
	if config, exists := c.Branches[branchName]; exists {
		return &config
	}

	for branchType, config := range c.Branches {
		if strings.HasPrefix(branchName, branchType+"/") {
			return &config
		}
	}

	return &BranchConfig{
		Increment: "Patch",
		Tag:       "{BranchName}",
		Regex:     "",
	}
}
