package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/VirtuallyScott/battle-tested-devops/gitversion-go/internal/version"
	"github.com/VirtuallyScott/battle-tested-devops/gitversion-go/pkg/gitversion"
)

const (
	Version    = "1.0.0"
	ScriptName = "gitversion"
)

func main() {
	var (
		help           = flag.Bool("h", false, "Show help message")
		helpLong       = flag.Bool("help", false, "Show help message")
		ver            = flag.Bool("v", false, "Show version information")
		versionLong    = flag.Bool("version", false, "Show version information")
		output         = flag.String("o", "text", "Output format (json|text|AssemblySemVer|AssemblySemFileVer)")
		outputLong     = flag.String("output", "text", "Output format (json|text|AssemblySemVer|AssemblySemFileVer)")
		configFile     = flag.String("c", "", "Path to configuration file")
		configFileLong = flag.String("config", "", "Path to configuration file")
		branch         = flag.String("b", "", "Target branch")
		branchLong     = flag.String("branch", "", "Target branch")
		workflow       = flag.String("w", "gitflow", "Workflow type (gitflow|githubflow|trunk)")
		workflowLong   = flag.String("workflow", "gitflow", "Workflow type (gitflow|githubflow|trunk)")
		major          = flag.Bool("major", false, "Force major version increment")
		minor          = flag.Bool("minor", false, "Force minor version increment")
		patch          = flag.Bool("patch", false, "Force patch version increment")
		nextVersion    = flag.String("next-version", "", "Override next version")
	)

	flag.Parse()

	if *help || *helpLong {
		showHelp()
		return
	}

	if *ver || *versionLong {
		showVersion()
		return
	}

	debug := os.Getenv("DEBUG") == "true"

	outputFormat := *output
	if *outputLong != "text" {
		outputFormat = *outputLong
	}

	configPath := *configFile
	if *configFileLong != "" {
		configPath = *configFileLong
	}

	targetBranch := *branch
	if *branchLong != "" {
		targetBranch = *branchLong
	}

	workflowType := *workflow
	if *workflowLong != "gitflow" {
		workflowType = *workflowLong
	}

	var forceIncrement string
	if *major {
		forceIncrement = "major"
	} else if *minor {
		forceIncrement = "minor"
	} else if *patch {
		forceIncrement = "patch"
	}

	opts := &gitversion.Options{
		OutputFormat:   gitversion.OutputFormat(outputFormat),
		ConfigFile:     configPath,
		TargetBranch:   targetBranch,
		Workflow:       version.WorkflowType(workflowType),
		ForceIncrement: forceIncrement,
		NextVersion:    *nextVersion,
		Debug:          debug,
	}

	gv, err := gitversion.New(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}

	result, err := gv.Calculate(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func showHelp() {
	fmt.Printf(`%s v%s - GitVersion Go implementation

USAGE:
    %s [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -v, --version           Show version information
    -o, --output FORMAT     Output format (json|text|AssemblySemVer|AssemblySemFileVer) [default: text]
    -c, --config FILE       Path to configuration file
    -b, --branch BRANCH     Target branch [default: current branch]
    -w, --workflow TYPE     Workflow type (gitflow|githubflow|trunk) [default: gitflow]
    --major                 Force major version increment
    --minor                 Force minor version increment
    --patch                 Force patch version increment
    --next-version VERSION  Override next version

EXAMPLES:
    %s                    # Calculate version for current branch
    %s -o json            # Output as JSON
    %s -o AssemblySemVer  # Output AssemblySemVer only
    %s -o AssemblySemFileVer # Output AssemblySemFileVer only
    %s -b main            # Calculate version for main branch
    %s --major            # Force major increment

ENVIRONMENT VARIABLES:
    DEBUG=true              Enable debug logging

`, ScriptName, Version, ScriptName, ScriptName, ScriptName, ScriptName, ScriptName, ScriptName, ScriptName)
}

func showVersion() {
	fmt.Printf("%s v%s\n", ScriptName, Version)
}