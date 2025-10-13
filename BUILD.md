# Building GitVersion-Go

This guide covers building, testing, and developing the GitVersion-Go implementation.

## Prerequisites

### Required

- **Go 1.21 or later** - [Install Go](https://golang.org/dl/)
- **Git 2.0 or later** - For version control operations
- **Make** - For build automation (optional but recommended)

### Optional

- **golangci-lint** - For code linting ([Installation guide](https://golangci-lint.run/usage/install/))
- **Docker** - For containerized builds

## Quick Start

```bash
# Clone the repository
git clone https://github.com/VirtuallyScott/battle-tested-devops.git
cd battle-tested-devops/gitversion-go

# Download dependencies
go mod tidy

# Build the binary
make build

# Run tests
make test

# Install locally
make install-system
```

## Build Commands

### Using Make (Recommended)

```bash
# Build for current platform
make build

# Build for all supported platforms
make build-all

# Clean build artifacts
make clean

# Show all available targets
make help
```

### Using Go Directly

```bash
# Simple build
go build -o gitversion ./cmd

# Build with version information
go build -ldflags "-X main.Version=$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')" -o gitversion ./cmd

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o gitversion-linux-amd64 ./cmd
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run tests with coverage
make test-coverage

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

### Test Structure

```
gitversion-go/
├── pkg/semver/version_test.go           # Semantic version parsing tests
├── pkg/config/config_test.go            # Configuration loading tests
├── pkg/gitversion/output_test.go        # Output formatting tests
├── internal/git/repository_test.go      # Git operations tests
├── internal/version/calculator_test.go  # Version calculation tests
└── tests/integration_test.go            # End-to-end integration tests
```

### Test Coverage

View test coverage in your browser:

```bash
make test-coverage
open coverage.html
```

## Code Quality

### Linting

```bash
# Run golangci-lint (requires installation)
make lint

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

### Formatting

```bash
# Format code
make fmt

# Verify formatting
go fmt ./...
```

### Static Analysis

```bash
# Run go vet
make vet

# Run all quality checks
make quality
```

## Development Workflow

### Setting Up Development Environment

```bash
# Clone and setup
git clone https://github.com/VirtuallyScott/battle-tested-devops.git
cd battle-tested-devops/gitversion-go

# Install dependencies
go mod tidy

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run development checks
make dev
```

### Development Commands

```bash
# Quick development workflow (format, vet, test)
make dev

# Watch for changes and run tests (requires entr)
find . -name "*.go" | entr -r make test-unit

# Build and test in one command
make build test
```

### Adding New Features

1. **Write tests first** - Add tests in appropriate `*_test.go` files
2. **Implement feature** - Add implementation code
3. **Run quality checks** - `make quality`
4. **Update documentation** - Update README.md if needed

## Installation

### Local Installation

```bash
# Install to GOPATH/bin
make install

# Install to /usr/local/bin (requires sudo)
make install-system

# Manual installation
cp build/gitversion /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/VirtuallyScott/battle-tested-devops/gitversion-go/cmd@latest
```

## Cross-Platform Builds

### Build for All Platforms

```bash
# Build for multiple platforms
make build-all

# Results in build/ directory:
# - gitversion-linux-amd64
# - gitversion-darwin-amd64  
# - gitversion-darwin-arm64
# - gitversion-windows-amd64.exe
```

### Manual Cross-Platform Builds

```bash
# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o gitversion-linux-amd64 ./cmd

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o gitversion-darwin-amd64 ./cmd

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gitversion-darwin-arm64 ./cmd

# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o gitversion-windows-amd64.exe ./cmd

# Linux (ARM64)
GOOS=linux GOARCH=arm64 go build -o gitversion-linux-arm64 ./cmd
```

## Docker Builds

### Using Docker

```bash
# Create Dockerfile
cat > Dockerfile << 'EOF'
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')" -o gitversion ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /root/
COPY --from=builder /app/gitversion .
ENTRYPOINT ["./gitversion"]
EOF

# Build Docker image
docker build -t gitversion-go .

# Run in container
docker run --rm -v $(pwd):/repo -w /repo gitversion-go
```

### Multi-stage Build for Size Optimization

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o gitversion ./cmd

FROM scratch
COPY --from=builder /app/gitversion /gitversion
ENTRYPOINT ["/gitversion"]
```

## Debugging

### Debug Mode

```bash
# Enable debug logging
DEBUG=true ./build/gitversion

# Debug with specific options
DEBUG=true ./build/gitversion --output json --branch develop
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug ./cmd -- --help

# Debug tests
dlv test ./pkg/gitversion
```

### Profiling

```bash
# CPU profiling
go build -o gitversion ./cmd
./gitversion --cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go build -o gitversion ./cmd  
./gitversion --memprofile=mem.prof
go tool pprof mem.prof
```

## Benchmarking

### Performance Testing

```bash
# Run benchmarks
go test -bench=. ./...

# Benchmark with memory allocation info
go test -bench=. -benchmem ./...

# Compare with shell implementation
time ./gitversion.sh
time ./build/gitversion
```

### Custom Benchmarks

```go
// Add to *_test.go files
func BenchmarkVersionCalculation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Your benchmark code
    }
}
```

## Troubleshooting

### Common Build Issues

#### Go Module Issues

```bash
# Clear module cache
go clean -modcache

# Update dependencies
go mod tidy
go mod download
```

#### Build Failures

```bash
# Check Go version
go version

# Verify GOPATH and GOROOT
go env

# Clean and rebuild
make clean
make build
```

#### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestSpecificFunction ./pkg/semver

# Run tests with race detection
go test -race ./...
```

### Platform-Specific Issues

#### Windows

```bash
# Use Git Bash or PowerShell
go build -o gitversion.exe ./cmd

# Set execution policy (PowerShell)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### macOS

```bash
# Install Xcode command line tools
xcode-select --install

# Install make via Homebrew
brew install make
```

#### Linux

```bash
# Install build essentials
sudo apt-get update
sudo apt-get install build-essential git

# Or on CentOS/RHEL
sudo yum groupinstall "Development Tools"
sudo yum install git
```

## Continuous Integration

### GitHub Actions

```yaml
name: Build and Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Test
      run: make test
    
    - name: Build
      run: make build-all
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: binaries
        path: build/
```

### GitLab CI

```yaml
stages:
  - test
  - build

variables:
  GO_VERSION: "1.21"

test:
  stage: test
  image: golang:${GO_VERSION}
  script:
    - go mod download
    - make test-coverage
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

build:
  stage: build
  image: golang:${GO_VERSION}
  script:
    - make build-all
  artifacts:
    paths:
      - build/
    expire_in: 1 week
```

## Release Process

### Creating Releases

```bash
# Tag a new release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build release binaries
make clean
make build-all

# Create release archive
tar -czf gitversion-go-v1.0.0.tar.gz build/

# Generate checksums
cd build/
sha256sum * > checksums.txt
```

### Automated Releases

Use GitHub Actions or similar CI/CD to automate releases:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: 1.21
    
    - name: Build
      run: make build-all
    
    - name: Create Release
      uses: goreleaser/goreleaser-action@v4
      with:
        version: latest
        args: release --rm-dist
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Contributing

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write clear, descriptive commit messages
- Add tests for new functionality
- Update documentation as needed

### Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run quality checks: `make quality`
5. Submit a pull request

For more details, see the main [README.md](README.md) file.