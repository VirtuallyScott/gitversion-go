package gitversion

import (
	"fmt"
	"os"

	"github.com/VirtuallyScott/gitversion-go/internal/git"
	"github.com/VirtuallyScott/gitversion-go/internal/version"
	"github.com/VirtuallyScott/gitversion-go/pkg/config"
)

type Options struct {
	OutputFormat   OutputFormat
	ConfigFile     string
	TargetBranch   string
	Workflow       version.WorkflowType
	ForceIncrement string
	NextVersion    string
	Debug          bool
}

type GitVersion struct {
	repo       *git.Repository
	config     *config.Config
	calculator *version.Calculator
	formatter  *Formatter
	debug      bool
}

func New(opts *Options) (*GitVersion, error) {
	repo := git.NewRepository()

	if !repo.IsRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	cfg, err := config.LoadConfig(opts.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	calculator := version.NewCalculator(repo, cfg)
	formatter := NewFormatter(repo)

	return &GitVersion{
		repo:       repo,
		config:     cfg,
		calculator: calculator,
		formatter:  formatter,
		debug:      opts.Debug,
	}, nil
}

func (gv *GitVersion) Calculate(opts *Options) (string, error) {
	branch := opts.TargetBranch
	if branch == "" {
		var err error
		branch, err = gv.repo.GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	if gv.debug {
		gv.logDebug("Target branch: %s", branch)
		gv.logDebug("Workflow: %s", opts.Workflow)
		gv.logDebug("Force increment: %s", opts.ForceIncrement)
		gv.logDebug("Next version: %s", opts.NextVersion)
		gv.logDebug("Config next version: %s", gv.config.NextVersion)
	}

	// Use config NextVersion if no command line override provided
	nextVersion := opts.NextVersion
	if nextVersion == "" && gv.config.NextVersion != "" {
		nextVersion = gv.config.NextVersion
		if gv.debug {
			gv.logDebug("Using config next version: %s", nextVersion)
		}
	}

	version, err := gv.calculator.CalculateVersion(branch, opts.Workflow, opts.ForceIncrement, nextVersion)
	if err != nil {
		return "", fmt.Errorf("failed to calculate version: %w", err)
	}

	if gv.debug {
		gv.logDebug("Calculated version: %s", version.String())
	}

	output, err := gv.formatter.Format(version, opts.OutputFormat, branch)
	if err != nil {
		return "", fmt.Errorf("failed to format output: %w", err)
	}

	return output, nil
}

func (gv *GitVersion) logDebug(format string, args ...interface{}) {
	if gv.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
