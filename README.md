# GitVersion Go

A high-performance Go implementation of [GitVersion](https://github.com/GitTools/GitVersion) that automatically generates semantic version numbers from your Git repository history. This implementation provides the same functionality as the original GitVersion tool and the shell implementation in gitversion-sh, but with the performance and portability benefits of Go.

## Features

- **Automatic Semantic Versioning**: Calculates version numbers based on Git history and branch structure
- **Multiple Workflow Support**: GitFlow, GitHubFlow, and trunk-based development workflows  
- **Commit Message Parsing**: Detects version increments from conventional commit messages and +semver tags
- **Branch-Aware Versioning**: Different versioning strategies for main, develop, feature, release, and hotfix branches
- **Flexible Output**: Support for text, JSON, AssemblySemVer, and AssemblySemFileVer formats
- **Pre-release Versions**: Automatic generation of alpha, beta, and feature-specific pre-release versions
- **Build Metadata**: Includes commit count and SHA information
- **Configuration Support**: JSON and YAML configuration files
- **High Performance**: Fast execution with minimal dependencies
- **Cross-platform**: Single binary for Linux, macOS, and Windows

## Installation

### Download Binary

Download the latest binary from the releases page or build from source:

```bash
# Download for Linux
curl -L -o gitversion https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-linux-amd64
chmod +x gitversion

# Download for macOS (Intel)
curl -L -o gitversion https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-darwin-amd64
chmod +x gitversion

# Download for macOS (Apple Silicon)
curl -L -o gitversion https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-darwin-arm64
chmod +x gitversion

# Move to PATH
sudo mv gitversion /usr/local/bin/
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/VirtuallyScott/gitversion-go.git
cd gitversion-go

# Build the binary
make build

# Install to /usr/local/bin
make install-system
```

### Go Install

```bash
go install github.com/VirtuallyScott/gitversion-go/cmd@latest
```

## Usage

### Basic Usage

```bash
# Calculate version for current branch
gitversion

# Output: 1.2.3+5+abc1234
```

### Command Line Options

```bash
gitversion [OPTIONS]

OPTIONS:
    -h, --help              Show help message
    -v, --version           Show version information
    -o, --output FORMAT     Output format (json|text|AssemblySemVer|AssemblySemFileVer) [default: text]
    -c, --config FILE       Path to configuration file  
    -b, --branch BRANCH     Target branch [default: current branch]
    -w, --workflow TYPE     Workflow type (gitflow|githubflow|trunk) [default: gitflow]
    --major                 Force major version increment
    --minor                 Force minor version increment  
    --patch                 Force patch version increment
    --next-version VERSION  Override next version
```

### Examples

```bash
# Basic version calculation
gitversion

# JSON output for CI/CD integration
gitversion --output json

# Output AssemblySemVer format (1.2.3.0)
gitversion --output AssemblySemVer

# Output AssemblySemFileVer format (1.2.3.0) 
gitversion --output AssemblySemFileVer

# Calculate version for specific branch
gitversion --branch main

# Force major version increment
gitversion --major

# Use GitHub Flow workflow
gitversion --workflow githubflow

# Override next version
gitversion --next-version 2.0.0

# Use configuration file
gitversion --config GitVersion.yml

# Enable debug logging
DEBUG=true gitversion
```

## Workflows

### GitFlow (Default)

Perfect for projects using the GitFlow branching model:

- **main/master**: Stable releases (1.0.0)
- **develop**: Development versions (1.1.0-alpha.5+10+abc1234)
- **feature/***: Feature branches (1.1.0-feature-name.3+5+def5678)
- **release/***: Release candidates (1.1.0-beta.2+8+ghi9012)
- **hotfix/***: Hotfix versions (1.0.1-hotfix.1+2+jkl3456)

### GitHubFlow

Simplified workflow for GitHub-style development:

- **main/master**: Stable releases
- **feature branches**: Pre-release versions with branch name

### Trunk-based

All branches treated as main branch versions.

## Configuration

### Configuration Files

GitVersion-go supports both JSON and YAML configuration files for advanced customization.

#### JSON Configuration (GitVersion.json)

```json
{
  "next-version": "1.0.0",
  "branches": {
    "main": {
      "increment": "Patch",
      "tag": "",
      "regex": "^master$|^main$"
    },
    "develop": {
      "increment": "Minor",
      "tag": "alpha",
      "regex": "^develop$"
    },
    "feature": {
      "increment": "Minor",
      "tag": "{BranchName}",
      "regex": "^features?[/-]"
    },
    "release": {
      "increment": "None",
      "tag": "beta",
      "regex": "^releases?[/-]"
    },
    "hotfix": {
      "increment": "Patch", 
      "tag": "hotfix",
      "regex": "^hotfix(es)?[/-]"
    }
  },
  "commit-message-incrementing": {
    "enabled": true,
    "increment-mode": "Enabled"
  }
}
```

#### YAML Configuration (GitVersion.yml)

```yaml
next-version: '1.0.0'

branches:
  main:
    increment: Patch
    tag: ''
    regex: '^master$|^main$'
    
  develop:
    increment: Minor
    tag: alpha
    regex: '^develop$'
    
  feature:
    increment: Minor
    tag: '{BranchName}'
    regex: '^features?[/-]'
    
  release:
    increment: None
    tag: beta
    regex: '^releases?[/-]'
    
  hotfix:
    increment: Patch
    tag: hotfix
    regex: '^hotfix(es)?[/-]'

commit-message-incrementing:
  enabled: true
  increment-mode: Enabled
```

### Configuration Usage

```bash
# Use JSON configuration
gitversion --config GitVersion.json

# Use YAML configuration  
gitversion --config GitVersion.yml

# Configuration with specific branch
gitversion --config GitVersion.yml --branch develop

# Override configuration with CLI args
gitversion --config GitVersion.yml --major --output json
```

## Version Increment Detection

The tool automatically detects version increments from commit messages:

### Semantic Version Tags

Add these tags to commit messages to control version increments:

```bash
git commit -m "fix: resolve login issue +semver: patch"
git commit -m "feat: add user profiles +semver: minor"
git commit -m "feat!: redesign API +semver: major"
```

### Conventional Commits

The tool also recognizes conventional commit patterns:

- `feat:` → Minor increment
- `feat!:` → Major increment (breaking change)
- `fix:` → Patch increment  
- `BREAKING CHANGE:` → Major increment

## Output Formats

### Text Output (Default)

```
1.2.3-alpha.5+10+abc1234
```

### Assembly Version Outputs

#### AssemblySemVer
```
1.2.3.0
```

#### AssemblySemFileVer
```
1.2.3.0
```

### JSON Output

```json
{
  "Major": 1,
  "Minor": 2,
  "Patch": 3,
  "PreReleaseTag": "alpha.5",
  "PreReleaseTagWithDash": "-alpha.5",
  "BuildMetaData": "10+abc1234",
  "BuildMetaDataPadded": "+10+abc1234",
  "FullBuildMetaData": "10+abc1234",
  "MajorMinorPatch": "1.2.3",
  "SemVer": "1.2.3-alpha.5+10+abc1234",
  "AssemblySemVer": "1.2.3.0",
  "AssemblySemFileVer": "1.2.3.0",
  "FullSemVer": "1.2.3-alpha.5+10+abc1234",
  "InformationalVersion": "1.2.3-alpha.5+10+abc1234",
  "BranchName": "develop",
  "EscapedBranchName": "develop",
  "Sha": "abc1234567890def",
  "ShortSha": "abc1234",
  "NuGetVersionV2": "1.2.3-alpha.5+10+abc1234",
  "NuGetVersion": "1.2.3-alpha.5+10+abc1234",
  "VersionSourceSha": "abc1234567890def",
  "CommitsSinceVersionSource": 10,
  "CommitDate": "2025-01-15 10:30:45 +0000"
}
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Build and Version
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for GitVersion
      
      - name: Setup GitVersion
        run: |
          curl -L -o gitversion https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-linux-amd64
          chmod +x gitversion
          sudo mv gitversion /usr/local/bin/
      
      - name: Calculate Version
        id: version
        run: |
          VERSION=$(gitversion)
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          
      - name: Build and Tag
        run: |
          echo "Building version: ${{ steps.version.outputs.version }}"
          docker build -t myapp:${{ steps.version.outputs.version }} .
```

### GitLab CI

```yaml
variables:
  GIT_DEPTH: 0  # Required for GitVersion

version:
  stage: version
  image: alpine/git
  before_script:
    - apk add --no-cache curl
    - curl -L -o gitversion https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-linux-amd64
    - chmod +x gitversion
  script:
    - VERSION=$(./gitversion)
    - echo "VERSION=$VERSION" >> build.env
  artifacts:
    reports:
      dotenv: build.env

build:
  stage: build
  script:
    - echo "Building version: $VERSION"
    - docker build -t myapp:$VERSION .
  dependencies:
    - version
```

### Jenkins

```groovy
pipeline {
    agent any
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Setup GitVersion') {
            steps {
                sh '''
                    curl -L -o gitversion https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-linux-amd64
                    chmod +x gitversion
                '''
            }
        }
        
        stage('Version') {
            steps {
                script {
                    def version = sh(script: './gitversion', returnStdout: true).trim()
                    env.VERSION = version
                    echo "Calculated version: ${version}"
                }
            }
        }
        
        stage('Build') {
            steps {
                sh "docker build -t myapp:${env.VERSION} ."
            }
        }
    }
}
```

### Azure DevOps

```yaml
trigger:
- main
- develop

pool:
  vmImage: 'ubuntu-latest'

variables:
  gitVersionPath: $(Agent.ToolsDirectory)/gitversion

steps:
- checkout: self
  fetchDepth: 0

- task: Bash@3
  displayName: 'Install GitVersion'
  inputs:
    targetType: 'inline'
    script: |
      curl -L -o $(gitVersionPath) https://github.com/VirtuallyScott/gitversion-go/releases/latest/download/gitversion-linux-amd64
      chmod +x $(gitVersionPath)

- task: Bash@3
  displayName: 'Calculate Version'
  inputs:
    targetType: 'inline'
    script: |
      VERSION=$($(gitVersionPath))
      echo "##vso[task.setvariable variable=VERSION]$VERSION"
      echo "Calculated version: $VERSION"

- task: Docker@2
  displayName: 'Build Docker Image'
  inputs:
    command: 'build'
    dockerfile: 'Dockerfile'
    tags: 'myapp:$(VERSION)'
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Development build with formatting and testing
make dev
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Vet code
make vet

# Run linter (requires golangci-lint)
make lint

# Run all quality checks
make quality
```

## Performance

GitVersion-go is designed for high performance:

- **Fast execution**: Typically completes in under 100ms
- **Low memory usage**: Minimal memory footprint
- **Single binary**: No runtime dependencies
- **Efficient Git operations**: Optimized Git command usage

### Benchmarks

```bash
# Shell implementation
time ./gitversion.sh
# real    0m0.450s

# Go implementation  
time ./gitversion
# real    0m0.045s
```

The Go implementation is approximately **10x faster** than the shell implementation.

## Architecture

### Project Structure

```
gitversion-go/
├── cmd/                    # Main application entry point
├── pkg/
│   ├── gitversion/        # Core GitVersion functionality
│   ├── config/            # Configuration management
│   └── semver/            # Semantic version parsing
├── internal/
│   ├── git/               # Git repository operations
│   └── version/           # Version calculation logic
├── tests/                 # Integration tests
└── Makefile              # Build and development tasks
```

### Key Components

- **cmd/main.go**: CLI interface and argument parsing
- **pkg/gitversion**: Core GitVersion functionality and output formatting
- **pkg/config**: Configuration file parsing (JSON/YAML)
- **pkg/semver**: Semantic version parsing and manipulation
- **internal/git**: Git repository operations and commit analysis
- **internal/version**: Version calculation logic and branch strategies

## Compatibility

### GitVersion Compatibility

This implementation strives for 100% compatibility with GitTools/GitVersion:

- **Same version calculation logic**
- **Identical JSON output format**
- **Compatible configuration files**
- **Same conventional commit parsing**
- **Matching branch strategy behavior**

### System Requirements

- **Operating Systems**: Linux, macOS, Windows
- **Architecture**: amd64, arm64
- **Git**: Version 2.0 or later
- **Dependencies**: None (static binary)

### Go Version

- **Minimum Go version**: 1.21
- **Tested with**: Go 1.21, 1.22

## Troubleshooting

### Debug Mode

Enable debug logging to see how versions are calculated:

```bash
DEBUG=true gitversion
```

### Common Issues

1. **Not a git repository**: Ensure you're running the command from within a Git repository
2. **No version tags found**: The tool starts from 0.0.0 if no semantic version tags exist
3. **Invalid tag format**: Ensure tags follow semantic versioning (v1.2.3 or 1.2.3)
4. **Permission denied**: Make sure the binary has execute permissions

### Validation

Test version calculation without making changes:

```bash
# Test different scenarios
gitversion --branch main
gitversion --branch develop
gitversion --major
gitversion --next-version 2.0.0
gitversion --output json | jq .
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run quality checks (`make quality`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Workflow

```bash
# Clone the repository
git clone https://github.com/VirtuallyScott/gitversion-go.git
cd gitversion-go

# Install dependencies
go mod tidy

# Make changes and test
make dev

# Build and test
make build
make test

# Run quality checks
make quality
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the original [GitVersion](https://github.com/GitTools/GitVersion) project
- Compatible with the shell implementation in [gitversion-sh](../gitversion-sh)
- Follows [Semantic Versioning](https://semver.org/) specifications
- Compatible with [Conventional Commits](https://www.conventionalcommits.org/)
- Built with ❤️ in Go

## Related Projects

- [GitTools/GitVersion](https://github.com/GitTools/GitVersion) - Original .NET implementation
- [gitversion-sh](../gitversion-sh) - Shell script implementation
- [Semantic Versioning](https://semver.org/) - Versioning specification
- [Conventional Commits](https://www.conventionalcommits.org/) - Commit message specification