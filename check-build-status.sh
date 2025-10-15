#!/usr/bin/env bash

#
# check-build-status.sh
# Build Status Checker for GitVersion-Go
#
# This script displays current build status and expected artifacts
# for the GitVersion-Go project.
#

set -euo pipefail

# Global variables
readonly SCRIPT_NAME="$(basename "$0")"
readonly REPO_OWNER="VirtuallyScott"
readonly REPO_NAME="gitversion-go"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

#
# Print colored output
#
print_info() {
    echo -e "${BLUE}$*${NC}"
}

print_success() {
    echo -e "${GREEN}$*${NC}"
}

print_warning() {
    echo -e "${YELLOW}$*${NC}"
}

print_error() {
    echo -e "${RED}$*${NC}" >&2
}

print_header() {
    echo -e "${CYAN}$*${NC}"
}

#
# Print usage information
#
usage() {
    cat << EOF
Usage: $SCRIPT_NAME [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    -v, --version       Show script version
    -q, --quiet         Show minimal output

DESCRIPTION:
    Display current build status and expected artifacts for GitVersion-Go project.
    This script shows the current branch, version, recent commits, and links to
    monitor GitHub Actions workflows and releases.

EXAMPLES:
    $SCRIPT_NAME                    # Show full build status
    $SCRIPT_NAME --quiet            # Show minimal status
    $SCRIPT_NAME --help             # Show this help

EOF
}

#
# Check if required commands exist
#
check_dependencies() {
    local missing_deps=()

    if ! command -v git >/dev/null 2>&1; then
        missing_deps+=("git")
    fi

    if ! command -v gitversion >/dev/null 2>&1; then
        missing_deps+=("gitversion")
    fi

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "❌ Missing required dependencies: ${missing_deps[*]}"
        print_info "Please install missing dependencies and try again"
        return 1
    fi

    return 0
}

#
# Get current git information
#
get_git_info() {
    local current_branch current_version last_commit

    if ! current_branch=$(git branch --show-current 2>/dev/null); then
        print_error "Failed to get current git branch"
        return 1
    fi

    if ! current_version=$(gitversion 2>/dev/null); then
        print_error "Failed to get version from gitversion"
        return 1
    fi

    if ! last_commit=$(git log -1 --pretty=format:'%h - %s (%cr)' 2>/dev/null); then
        print_error "Failed to get last commit information"
        return 1
    fi

    # Remove any trailing characters (like %)
    current_version=$(echo "$current_version" | tr -d '%')

    echo "$current_branch|$current_version|$last_commit"
}

#
# Display expected artifacts
#
show_expected_artifacts() {
    local version="$1"

    print_header "📦 Expected artifacts for version $version:"
    echo "  - gitversion-linux-amd64"
    echo "  - gitversion-linux-arm64"
    echo "  - gitversion-darwin-amd64"
    echo "  - gitversion-darwin-arm64"
    echo "  - gitversion-windows-amd64.exe"
    echo "  - gitversion-windows-arm64.exe"
    echo "  - SHA256 checksums for all binaries"
}

#
# Display workflow expectations
#
show_workflow_expectations() {
    local version="$1"

    print_header "💡 The workflow should:"
    echo "  1. Calculate version: $version"
    echo "  2. Build all platform binaries"
    echo "  3. Run tests with coverage"
    echo "  4. Create GitHub release with artifacts"
    echo "  5. Tag repository with v$version"
}

#
# Display build status (full version)
#
show_full_status() {
    local branch="$1"
    local version="$2"
    local last_commit="$3"

    echo
    print_header "🚀 GitVersion-Go Build Status Check"
    print_header "=================================="
    echo
    print_info "📍 Current Branch: $branch"
    print_info "🏷️  Current Version: $version"
    print_info "📅 Last Commit: $last_commit"
    echo
    print_success "🔗 GitHub Actions: https://github.com/$REPO_OWNER/$REPO_NAME/actions"
    print_success "📦 Releases: https://github.com/$REPO_OWNER/$REPO_NAME/releases"
    echo

    show_expected_artifacts "$version"
    echo
    show_workflow_expectations "$version"
}

#
# Display build status (quiet version)
#
show_quiet_status() {
    local branch="$1"
    local version="$2"

    echo "Branch: $branch | Version: $version"
}

#
# Main function
#
main() {
    local quiet_mode="false"

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--version)
                echo "$SCRIPT_NAME version 1.0.0"
                exit 0
                ;;
            -q|--quiet)
                quiet_mode="true"
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
        shift
    done

    # Check dependencies
    if ! check_dependencies; then
        exit 1
    fi

    # Get git information
    local git_info
    if ! git_info=$(get_git_info); then
        exit 1
    fi

    local current_branch current_version last_commit
    IFS='|' read -r current_branch current_version last_commit <<< "$git_info"

    # Show status based on mode
    if [[ "$quiet_mode" == "true" ]]; then
        show_quiet_status "$current_branch" "$current_version"
    else
        show_full_status "$current_branch" "$current_version" "$last_commit"
    fi
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
