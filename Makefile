.PHONY: build test test-unit test-integration clean install lint fmt vet quality dev help
.PHONY: pre-commit pre-commit-install pre-commit-update
.PHONY: git-status git-sync git-feature-start git-feature-finish git-release-start git-release-finish
.PHONY: git-hotfix-start git-hotfix-finish git-merge-to-develop git-merge-to-main version-info
.PHONY: release-list release-delete tag-list tag-delete tag-rename github-trigger

# Build variables
BINARY_NAME=gitversion
BUILD_DIR=build
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GITVERSION_VERSION=$(shell gitversion -c gitversion.yml -o json 2>/dev/null | jq -r '.MajorMinorPatch + "-" + .PreReleaseTag' 2>/dev/null || echo "1.0.0")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Git flow variables
CURRENT_BRANCH=$(shell git branch --show-current)
DEFAULT_BRANCH=main
DEVELOP_BRANCH=develop

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
all: build

# =============================================================================
# BUILD TARGETS
# =============================================================================

# Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./gitversion

# Build for multiple platforms
build-all:
	@echo "$(GREEN)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./gitversion
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./gitversion
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./gitversion
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./gitversion
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./gitversion
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./gitversion

# =============================================================================
# TEST TARGETS
# =============================================================================

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	go test -v -short ./...

# Run integration tests
test-integration: build
	@echo "$(GREEN)Running integration tests...$(NC)"
	go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# =============================================================================
# CODE QUALITY TARGETS
# =============================================================================

# Format the code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...

# Vet the code
vet:
	@echo "$(GREEN)Vetting code...$(NC)"
	go vet ./...

# Lint the code
lint:
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

# Run all quality checks
quality: fmt vet lint test-unit

# =============================================================================
# PRE-COMMIT TARGETS
# =============================================================================

# Install pre-commit hooks
pre-commit-install:
	@echo "$(GREEN)Installing pre-commit hooks...$(NC)"
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install; \
		echo "$(GREEN)Pre-commit hooks installed successfully$(NC)"; \
	else \
		echo "$(RED)pre-commit not installed. Install with: pip install pre-commit$(NC)"; \
		exit 1; \
	fi

# Update pre-commit hooks
pre-commit-update:
	@echo "$(GREEN)Updating pre-commit hooks...$(NC)"
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit autoupdate; \
	else \
		echo "$(RED)pre-commit not installed. Run 'make pre-commit-install' first$(NC)"; \
		exit 1; \
	fi

# Run pre-commit on all files
pre-commit:
	@echo "$(GREEN)Running pre-commit on all files...$(NC)"
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --all-files; \
	else \
		echo "$(RED)pre-commit not installed. Run 'make pre-commit-install' first$(NC)"; \
		exit 1; \
	fi

# =============================================================================
# GIT FLOW TARGETS
# =============================================================================

# Show current git status
git-status:
	@echo "$(BLUE)Current branch: $(CURRENT_BRANCH)$(NC)"
	@echo "$(BLUE)Git status:$(NC)"
	@git status --short
	@echo ""
	@echo "$(BLUE)Local branches:$(NC)"
	@git branch
	@echo ""
	@echo "$(BLUE)Remote branches:$(NC)"
	@git branch -r

# Show GitVersion information
version-info:
	@echo "$(BLUE)GitVersion Information:$(NC)"
	@echo "$(GREEN)Current MajorMinorPatch version: $(GITVERSION_VERSION)$(NC)"
	@echo "$(YELLOW)Full GitVersion output:$(NC)"
	@gitversion -c gitversion.yml
	@echo ""
	@echo "$(YELLOW)JSON format:$(NC)"
	@gitversion -c gitversion.yml -o json

# Sync current branch with remote
git-sync:
	@echo "$(GREEN)Syncing $(CURRENT_BRANCH) with remote...$(NC)"
	git fetch origin
	@if git ls-remote --heads origin $(CURRENT_BRANCH) | grep -q $(CURRENT_BRANCH); then \
		git pull origin $(CURRENT_BRANCH); \
		echo "$(GREEN)Synced with origin/$(CURRENT_BRANCH)$(NC)"; \
	else \
		echo "$(YELLOW)Branch $(CURRENT_BRANCH) doesn't exist on remote$(NC)"; \
	fi

# Start a new feature branch from develop
git-feature-start:
	@if [ -z "$(FEATURE)" ]; then \
		echo "$(RED)Usage: make git-feature-start FEATURE=feature-name$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating feature branch: feature/$(FEATURE)$(NC)"
	git checkout $(DEVELOP_BRANCH)
	git pull origin $(DEVELOP_BRANCH)
	git checkout -b feature/$(FEATURE)
	git push -u origin feature/$(FEATURE)
	@echo "$(GREEN)Feature branch feature/$(FEATURE) created and pushed$(NC)"

# Finish a feature branch (merge to develop)
git-feature-finish:
	@if [ -z "$(FEATURE)" ]; then \
		echo "$(RED)Usage: make git-feature-finish FEATURE=feature-name$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Finishing feature branch: feature/$(FEATURE)$(NC)"
	@if [ "$(CURRENT_BRANCH)" != "feature/$(FEATURE)" ]; then \
		git checkout feature/$(FEATURE); \
	fi
	git pull origin feature/$(FEATURE)
	make quality
	git checkout $(DEVELOP_BRANCH)
	git pull origin $(DEVELOP_BRANCH)
	git merge --no-ff feature/$(FEATURE) -m "Merge feature/$(FEATURE) into $(DEVELOP_BRANCH)"
	git push origin $(DEVELOP_BRANCH)
	@echo "$(GREEN)Feature branch merged to $(DEVELOP_BRANCH)$(NC)"
	@echo "$(YELLOW)To delete the feature branch, run:$(NC)"
	@echo "  git branch -d feature/$(FEATURE)"
	@echo "  git push origin --delete feature/$(FEATURE)"

# Start a release branch from main (merges current branch first if needed)
git-release-start:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(YELLOW)No VERSION specified, using GitVersion to calculate next version...$(NC)"; \
		RELEASE_VERSION=$(GITVERSION_VERSION); \
	else \
		RELEASE_VERSION=$(VERSION); \
	fi; \
	echo "$(GREEN)Creating release branch: release/$$RELEASE_VERSION$(NC)"; \
	echo "$(BLUE)Current branch: $(CURRENT_BRANCH)$(NC)"; \
	if [ "$(CURRENT_BRANCH)" != "$(DEFAULT_BRANCH)" ]; then \
		echo "$(YELLOW)Merging $(CURRENT_BRANCH) into $(DEFAULT_BRANCH) first...$(NC)"; \
		make quality; \
		git checkout $(DEFAULT_BRANCH); \
		git pull origin $(DEFAULT_BRANCH); \
		git merge --no-ff $(CURRENT_BRANCH) -m "Merge $(CURRENT_BRANCH) into $(DEFAULT_BRANCH) for release"; \
		git push origin $(DEFAULT_BRANCH); \
		echo "$(GREEN)Successfully merged $(CURRENT_BRANCH) into $(DEFAULT_BRANCH)$(NC)"; \
	else \
		echo "$(BLUE)Already on $(DEFAULT_BRANCH), pulling latest...$(NC)"; \
		git pull origin $(DEFAULT_BRANCH); \
	fi; \
	git checkout -b release/$$RELEASE_VERSION; \
	echo "$(YELLOW)Updating version in gitversion.yml...$(NC)"; \
	if [ -f "gitversion.yml" ]; then \
		sed -i.bak "s/next-version: .*/next-version: $$RELEASE_VERSION/" gitversion.yml && rm gitversion.yml.bak; \
	fi; \
	git add gitversion.yml; \
	git commit -m "Bump version to $$RELEASE_VERSION for release"; \
	git push -u origin release/$$RELEASE_VERSION; \
	echo "$(GREEN)Release branch release/$$RELEASE_VERSION created and pushed$(NC)"; \
	echo "$(BLUE)This will trigger the GitHub Actions release build$(NC)"

# Finish a release branch (merge to main and develop, create tag)
git-release-finish:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(YELLOW)No VERSION specified, using GitVersion to calculate version...$(NC)"; \
		RELEASE_VERSION=$(GITVERSION_VERSION); \
	else \
		RELEASE_VERSION=$(VERSION); \
	fi; \
	echo "$(GREEN)Finishing release with version: $$RELEASE_VERSION$(NC)"; \
	if [ "$(CURRENT_BRANCH)" != "release/$$RELEASE_VERSION" ]; then \
		echo "$(YELLOW)Switching to release branch: release/$$RELEASE_VERSION$(NC)"; \
		git checkout release/$$RELEASE_VERSION; \
	fi; \
	git pull origin release/$$RELEASE_VERSION; \
	make quality; \
	echo "$(GREEN)Merging to $(DEFAULT_BRANCH)...$(NC)"; \
	git checkout $(DEFAULT_BRANCH); \
	git pull origin $(DEFAULT_BRANCH); \
	git merge --no-ff release/$$RELEASE_VERSION -m "Merge release/$$RELEASE_VERSION into $(DEFAULT_BRANCH)"; \
	git tag -a v$$RELEASE_VERSION -m "Release version $$RELEASE_VERSION"; \
	git push origin $(DEFAULT_BRANCH); \
	git push origin v$$RELEASE_VERSION; \
	echo "$(GREEN)Merging back to $(DEVELOP_BRANCH)...$(NC)"; \
	git checkout $(DEVELOP_BRANCH); \
	git pull origin $(DEVELOP_BRANCH); \
	git merge --no-ff $(DEFAULT_BRANCH) -m "Merge $(DEFAULT_BRANCH) back into $(DEVELOP_BRANCH)"; \
	git push origin $(DEVELOP_BRANCH); \
	echo "$(GREEN)Release $$RELEASE_VERSION finished and tagged as v$$RELEASE_VERSION$(NC)"; \
	echo "$(YELLOW)To delete the release branch, run:$(NC)"; \
	echo "  git branch -d release/$$RELEASE_VERSION"; \
	echo "  git push origin --delete release/$$RELEASE_VERSION"

# Start a hotfix branch from main
git-hotfix-start:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Usage: make git-hotfix-start VERSION=x.y.z$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating hotfix branch: hotfix/$(VERSION)$(NC)"
	git checkout $(DEFAULT_BRANCH)
	git pull origin $(DEFAULT_BRANCH)
	git checkout -b hotfix/$(VERSION)
	git push -u origin hotfix/$(VERSION)
	@echo "$(GREEN)Hotfix branch hotfix/$(VERSION) created and pushed$(NC)"

# Finish a hotfix branch (merge to main and develop, create tag)
git-hotfix-finish:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Usage: make git-hotfix-finish VERSION=x.y.z$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Finishing hotfix branch: hotfix/$(VERSION)$(NC)"
	@if [ "$(CURRENT_BRANCH)" != "hotfix/$(VERSION)" ]; then \
		git checkout hotfix/$(VERSION); \
	fi
	git pull origin hotfix/$(VERSION)
	make quality
	# Merge to main
	git checkout $(DEFAULT_BRANCH)
	git pull origin $(DEFAULT_BRANCH)
	git merge --no-ff hotfix/$(VERSION) -m "Merge hotfix/$(VERSION) into $(DEFAULT_BRANCH)"
	git tag -a v$(VERSION) -m "Hotfix version $(VERSION)"
	git push origin $(DEFAULT_BRANCH)
	git push origin v$(VERSION)
	# Merge back to develop
	git checkout $(DEVELOP_BRANCH)
	git pull origin $(DEVELOP_BRANCH)
	git merge --no-ff $(DEFAULT_BRANCH) -m "Merge $(DEFAULT_BRANCH) back into $(DEVELOP_BRANCH)"
	git push origin $(DEVELOP_BRANCH)
	@echo "$(GREEN)Hotfix $(VERSION) finished and tagged$(NC)"
	@echo "$(YELLOW)To delete the hotfix branch, run:$(NC)"
	@echo "  git branch -d hotfix/$(VERSION)"
	@echo "  git push origin --delete hotfix/$(VERSION)"

# Merge current branch to develop
git-merge-to-develop:
	@echo "$(GREEN)Merging $(CURRENT_BRANCH) to $(DEVELOP_BRANCH)...$(NC)"
	@if [ "$(CURRENT_BRANCH)" = "$(DEVELOP_BRANCH)" ]; then \
		echo "$(YELLOW)Already on $(DEVELOP_BRANCH) branch$(NC)"; \
		exit 1; \
	fi
	make quality
	git checkout $(DEVELOP_BRANCH)
	git pull origin $(DEVELOP_BRANCH)
	git merge --no-ff $(CURRENT_BRANCH) -m "Merge $(CURRENT_BRANCH) into $(DEVELOP_BRANCH)"
	git push origin $(DEVELOP_BRANCH)
	@echo "$(GREEN)Successfully merged $(CURRENT_BRANCH) to $(DEVELOP_BRANCH)$(NC)"

# Merge current branch to main (use with caution)
git-merge-to-main:
	@echo "$(YELLOW)WARNING: Merging $(CURRENT_BRANCH) directly to $(DEFAULT_BRANCH)$(NC)"
	@echo "$(YELLOW)This should typically only be done for hotfixes$(NC)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	make quality
	git checkout $(DEFAULT_BRANCH)
	git pull origin $(DEFAULT_BRANCH)
	git merge --no-ff $(CURRENT_BRANCH) -m "Merge $(CURRENT_BRANCH) into $(DEFAULT_BRANCH)"
	git push origin $(DEFAULT_BRANCH)
	@echo "$(GREEN)Successfully merged $(CURRENT_BRANCH) to $(DEFAULT_BRANCH)$(NC)"

# =============================================================================
# INSTALLATION TARGETS
# =============================================================================

# Install the binary to GOPATH/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Install to /usr/local/bin (requires sudo)
install-system: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# =============================================================================
# RELEASE & TAG MANAGEMENT TARGETS
# =============================================================================

# List all GitHub releases
release-list:
	@echo "$(GREEN)Listing GitHub releases...$(NC)"
	@if command -v gh >/dev/null 2>&1; then \
		gh release list; \
	else \
		echo "$(RED)GitHub CLI (gh) not installed. Install with: brew install gh$(NC)"; \
		exit 1; \
	fi

# Delete a GitHub release
release-delete:
	@if [ -z "$(TAG)" ]; then \
		echo "$(RED)Usage: make release-delete TAG=tag-name$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Deleting GitHub release: $(TAG)$(NC)"
	@if command -v gh >/dev/null 2>&1; then \
		gh release delete "$(TAG)" --yes; \
		echo "$(GREEN)Deleted release $(TAG)$(NC)"; \
	else \
		echo "$(RED)GitHub CLI (gh) not installed. Install with: brew install gh$(NC)"; \
		exit 1; \
	fi

# List all git tags
tag-list:
	@echo "$(GREEN)Listing git tags...$(NC)"
	@echo "$(BLUE)Local tags:$(NC)"
	@git tag | sort -V
	@echo ""
	@echo "$(BLUE)Remote tags:$(NC)"
	@git ls-remote --tags origin | sed 's/.*refs\/tags\///' | sort -V

# Delete a git tag (local and remote)
tag-delete:
	@if [ -z "$(TAG)" ]; then \
		echo "$(RED)Usage: make tag-delete TAG=tag-name$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Deleting tag: $(TAG)$(NC)"
	@if git tag | grep -q "^$(TAG)$$"; then \
		git tag -d "$(TAG)"; \
		echo "$(GREEN)Deleted local tag $(TAG)$(NC)"; \
	else \
		echo "$(YELLOW)Local tag $(TAG) not found$(NC)"; \
	fi
	@if git ls-remote --tags origin | grep -q "refs/tags/$(TAG)$$"; then \
		git push origin --delete "$(TAG)"; \
		echo "$(GREEN)Deleted remote tag $(TAG)$(NC)"; \
	else \
		echo "$(YELLOW)Remote tag $(TAG) not found$(NC)"; \
	fi

# Rename a git tag (local and remote)
tag-rename:
	@if [ -z "$(OLD_TAG)" ] || [ -z "$(NEW_TAG)" ]; then \
		echo "$(RED)Usage: make tag-rename OLD_TAG=old-name NEW_TAG=new-name$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Renaming tag $(OLD_TAG) to $(NEW_TAG)...$(NC)"
	@if git tag | grep -q "^$(OLD_TAG)$$"; then \
		git tag $(NEW_TAG) $(OLD_TAG); \
		git tag -d $(OLD_TAG); \
		git push origin $(NEW_TAG); \
		git push origin --delete $(OLD_TAG); \
		echo "$(GREEN)Successfully renamed tag $(OLD_TAG) to $(NEW_TAG)$(NC)"; \
	else \
		echo "$(RED)Tag $(OLD_TAG) not found locally$(NC)"; \
		exit 1; \
	fi

# Trigger GitHub Actions workflow manually
github-trigger:
	@echo "$(GREEN)Manual GitHub Actions Trigger Helper$(NC)"
	@echo "======================================"
	@echo ""
	@CURRENT_VERSION=$$(./gitversion -c gitversion.yml 2>/dev/null | tr -d '%' || echo "0.0.1"); \
	echo "$(BLUE)Current Branch:$(NC) $(CURRENT_BRANCH)"; \
	echo "$(BLUE)Current Version:$(NC) $$CURRENT_VERSION"; \
	echo ""; \
	echo "$(GREEN)To manually trigger GitHub Actions:$(NC)"; \
	echo ""; \
	if command -v gh >/dev/null 2>&1; then \
		echo "$(YELLOW)Using GitHub CLI:$(NC)"; \
		echo "gh workflow run build-and-release.yml \\"; \
		echo "  --ref $(CURRENT_BRANCH) \\"; \
		echo "  -f version=\"$$CURRENT_VERSION\" \\"; \
		echo "  -f force_release=true"; \
		echo ""; \
		echo "$(YELLOW)Or via web interface:$(NC)"; \
	else \
		echo "$(YELLOW)Via web interface (GitHub CLI not available):$(NC)"; \
	fi; \
	echo "1. Go to: https://github.com/VirtuallyScott/gitversion-go/actions/workflows/build-and-release.yml"; \
	echo "2. Click 'Run workflow'"; \
	echo "3. Select branch: $(CURRENT_BRANCH)"; \
	echo "4. Enter version: $$CURRENT_VERSION"; \
	echo "5. Check 'Force create release'"; \
	echo "6. Click 'Run workflow'"; \
	echo ""

# =============================================================================
# UTILITY TARGETS
# =============================================================================

# Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Development workflow (quick checks)
dev: fmt vet test-unit

# Complete development workflow with pre-commit
dev-full: pre-commit quality

# =============================================================================
# HELP TARGET
# =============================================================================

# Help target
help:
	@echo "$(BLUE)GitVersion-Go Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Build Targets:$(NC)"
	@echo "  build           - Build the binary"
	@echo "  build-all       - Build for multiple platforms"
	@echo "  install         - Install to GOPATH/bin"
	@echo "  install-system  - Install to /usr/local/bin (requires sudo)"
	@echo ""
	@echo "$(YELLOW)Test Targets:$(NC)"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo ""
	@echo "$(YELLOW)Code Quality Targets:$(NC)"
	@echo "  fmt             - Format code"
	@echo "  vet             - Vet code"
	@echo "  lint            - Run golangci-lint"
	@echo "  quality         - Run all quality checks"
	@echo "  dev             - Quick development workflow (fmt, vet, test-unit)"
	@echo "  dev-full        - Complete development workflow (pre-commit + quality)"
	@echo ""
	@echo "$(YELLOW)Pre-commit Targets:$(NC)"
	@echo "  pre-commit-install - Install pre-commit hooks"
	@echo "  pre-commit-update  - Update pre-commit hooks"
	@echo "  pre-commit         - Run pre-commit on all files"
	@echo ""
	@echo "$(YELLOW)Git Flow Targets:$(NC)"
	@echo "  git-status         - Show current git status and branches"
	@echo "  version-info       - Show GitVersion information and calculated versions"
	@echo "  git-sync           - Sync current branch with remote"
	@echo "  git-feature-start  - Start new feature branch (FEATURE=name)"
	@echo "  git-feature-finish - Finish feature branch (FEATURE=name)"
	@echo "  git-release-start  - Start release branch (VERSION=x.y.z)"
	@echo "  git-release-finish - Finish release branch (VERSION=x.y.z)"
	@echo "  git-hotfix-start   - Start hotfix branch (VERSION=x.y.z)"
	@echo "  git-hotfix-finish  - Finish hotfix branch (VERSION=x.y.z)"
	@echo "  git-merge-to-develop - Merge current branch to develop"
	@echo "  git-merge-to-main    - Merge current branch to main (use with caution)"
	@echo ""
	@echo "$(YELLOW)Release & Tag Management:$(NC)"
	@echo "  release-list        - List all GitHub releases"
	@echo "  release-delete      - Delete GitHub release (TAG=name)"
	@echo "  tag-list           - List all git tags (local and remote)"
	@echo "  tag-delete         - Delete git tag (TAG=name)"
	@echo "  tag-rename         - Rename git tag (OLD_TAG=old NEW_TAG=new)"
	@echo "  github-trigger     - Show manual GitHub Actions trigger instructions"
	@echo ""
	@echo "$(YELLOW)Utility Targets:$(NC)"
	@echo "  clean           - Clean build artifacts"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make git-feature-start FEATURE=awesome-feature"
	@echo "  make git-feature-finish FEATURE=awesome-feature"
	@echo "  make git-release-start VERSION=1.2.0"
	@echo "  make git-release-finish VERSION=1.2.0"
	@echo "  make release-delete TAG=v1.0.0"
	@echo "  make tag-delete TAG=v0.0.1-beta"
	@echo "  make tag-rename OLD_TAG=v1.0.0-beta.37+37+861b6f2 NEW_TAG=v1.0.0-beta.37"
