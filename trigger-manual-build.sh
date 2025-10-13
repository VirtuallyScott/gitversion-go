#!/usr/bin/env bash

#
# trigger-manual-build.sh
# Manual GitHub Actions Trigger Helper for GitVersion-Go
#
# This script helps trigger GitHub Actions workflows manually with version override
# to solve the chicken-and-egg problem during initial builds.
#

set -euo pipefail

# Global variables
readonly SCRIPT_NAME="$(basename "$0")"
readonly REPO_OWNER="VirtuallyScott"
readonly REPO_NAME="gitversion-go"
readonly WORKFLOW_FILE="build-and-release.yml"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

#
# Print colored output
#
print_info() {
    echo -e "${BLUE}ℹ️  $*${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $*${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $*${NC}"
}

print_error() {
    echo -e "${RED}❌ $*${NC}" >&2
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
    -f, --force         Force create release (default: true)
    --no-force          Don't force create release

DESCRIPTION:
    This script helps you manually trigger GitHub Actions workflow with version override
    to bypass the chicken-and-egg problem where GitVersion needs to build itself.

EXAMPLES:
    $SCRIPT_NAME                    # Show manual trigger instructions
    $SCRIPT_NAME --no-force         # Show instructions without forcing release
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
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_info "Please install missing dependencies and try again"
        return 1
    fi
    
    return 0
}

#
# Get current git information
#
get_git_info() {
    local current_branch current_version
    
    if ! current_branch=$(git branch --show-current 2>/dev/null); then
        print_error "Failed to get current git branch"
        print_info "Make sure you're in a git repository"
        return 1
    fi
    
    if ! current_version=$(gitversion 2>/dev/null); then
        print_error "Failed to get version from gitversion"
        print_info "Make sure gitversion is properly installed and working"
        return 1
    fi
    
    # Remove any trailing characters (like %)
    current_version=$(echo "$current_version" | tr -d '%')
    
    echo "$current_branch|$current_version"
}

#
# Display manual trigger instructions
#
show_manual_instructions() {
    local branch="$1"
    local version="$2"
    local force_release="$3"
    
    echo
    print_info "🚀 Manual GitHub Actions Trigger Helper"
    echo "======================================"
    echo
    print_info "📍 Current Branch: $branch"
    print_info "🏷️  Current Version: $version"
    echo
    
    print_success "To manually trigger GitHub Actions with this version:"
    echo
    echo "1. Go to: https://github.com/$REPO_OWNER/$REPO_NAME/actions/workflows/$WORKFLOW_FILE"
    echo "2. Click 'Run workflow'"
    echo "3. Select branch: $branch"
    echo "4. Enter version: $version"
    echo "5. $([ "$force_release" = "true" ] && echo "Check" || echo "Uncheck") 'Force create release'"
    echo "6. Click 'Run workflow'"
    echo
    
    if command -v gh >/dev/null 2>&1; then
        print_success "Or use GitHub CLI:"
        echo
        echo "gh workflow run $WORKFLOW_FILE \\"
        echo "  --ref $branch \\"
        echo "  -f version=\"$version\" \\"
        echo "  -f force_release=$force_release"
        echo
    else
        print_warning "GitHub CLI (gh) not found. Install it for command-line workflow triggering."
    fi
    
    print_info "💡 This bypasses the chicken-and-egg problem by using your local version!"
}

#
# Main function
#
main() {
    local force_release="true"
    
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
            -f|--force)
                force_release="true"
                ;;
            --no-force)
                force_release="false"
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
    
    local current_branch current_version
    IFS='|' read -r current_branch current_version <<< "$git_info"
    
    # Show instructions
    show_manual_instructions "$current_branch" "$current_version" "$force_release"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi