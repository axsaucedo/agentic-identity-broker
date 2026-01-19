#!/usr/bin/env bash
# Common functions and variables for all scripts

# Get repository root, with fallback for non-git repositories
get_repo_root() {
    if git rev-parse --show-toplevel >/dev/null 2>&1; then
        git rev-parse --show-toplevel
    else
        # Fall back to script location for non-git repos
        local script_dir="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        (cd "$script_dir/../../.." && pwd)
    fi
}

# Get current branch, with fallback for non-git repositories
get_current_branch() {
    # First check if SPECIFY_FEATURE environment variable is set
    if [[ -n "${SPECIFY_FEATURE:-}" ]]; then
        echo "$SPECIFY_FEATURE"
        return
    fi

    # Then check git if available
    if git rev-parse --abbrev-ref HEAD >/dev/null 2>&1; then
        git rev-parse --abbrev-ref HEAD
        return
    fi

    # For non-git repos, try to find the latest feature directory
    local repo_root=$(get_repo_root)
    local specs_dir="$repo_root/specs"

    if [[ -d "$specs_dir" ]]; then
        local latest_feature=""
        local highest=0

        for dir in "$specs_dir"/*; do
            if [[ -d "$dir" ]]; then
                local dirname=$(basename "$dir")
                if [[ "$dirname" =~ ^([0-9]{3})- ]]; then
                    local number=${BASH_REMATCH[1]}
                    number=$((10#$number))
                    if [[ "$number" -gt "$highest" ]]; then
                        highest=$number
                        latest_feature=$dirname
                    fi
                fi
            fi
        done

        if [[ -n "$latest_feature" ]]; then
            echo "$latest_feature"
            return
        fi
    fi

    echo "main"  # Final fallback
}

# Check if we have git available
has_git() {
    git rev-parse --show-toplevel >/dev/null 2>&1
}

# Get base feature branch from environment variables
# This is useful when working on a non-standard branch that targets a feature branch
get_base_feature_branch() {
    # Check common CI/agent environment variables for base branch information
    local base_ref=""
    
    # GitHub Copilot agent sets COPILOT_AGENT_BASE_COMMIT
    if [[ -n "${COPILOT_AGENT_BASE_COMMIT:-}" ]]; then
        base_ref="$COPILOT_AGENT_BASE_COMMIT"
    # GitHub Actions sets GITHUB_BASE_REF for pull requests
    elif [[ -n "${GITHUB_BASE_REF:-}" ]]; then
        base_ref="$GITHUB_BASE_REF"
    fi
    
    # Extract branch name from refs (e.g., refs/heads/011-feature -> 011-feature)
    if [[ -n "$base_ref" ]]; then
        base_ref="${base_ref#refs/heads/}"
        base_ref="${base_ref#refs/remotes/origin/}"
        echo "$base_ref"
    fi
}

check_feature_branch() {
    local branch="$1"
    local has_git_repo="$2"

    # For non-git repos, we can't enforce branch naming but still provide output
    if [[ "$has_git_repo" != "true" ]]; then
        echo "[specify] Warning: Git repository not detected; skipped branch validation" >&2
        return 0
    fi

    # Check if current branch matches the feature branch pattern
    if [[ "$branch" =~ ^[0-9]{3}- ]]; then
        return 0
    fi
    
    # If current branch doesn't match, check if we have a base feature branch from environment
    local base_branch=$(get_base_feature_branch)
    if [[ -n "$base_branch" ]] && [[ "$base_branch" =~ ^[0-9]{3}- ]]; then
        echo "[specify] Info: Current branch '$branch' doesn't match feature pattern, but base branch '$base_branch' does" >&2
        return 0
    fi
    
    # Neither current nor base branch match the pattern
    echo "ERROR: Not on a feature branch. Current branch: $branch" >&2
    echo "Feature branches should be named like: 001-feature-name" >&2
    return 1
}

get_feature_dir() { echo "$1/specs/$2"; }

# Find feature directory by numeric prefix instead of exact branch match
# This allows multiple branches to work on the same spec (e.g., 004-fix-bug, 004-add-feature)
find_feature_dir_by_prefix() {
    local repo_root="$1"
    local branch_name="$2"
    local specs_dir="$repo_root/specs"

    # Extract numeric prefix from branch (e.g., "004" from "004-whatever")
    if [[ ! "$branch_name" =~ ^([0-9]{3})- ]]; then
        # If current branch doesn't have numeric prefix, try base branch
        local base_branch=$(get_base_feature_branch)
        if [[ -n "$base_branch" ]] && [[ "$base_branch" =~ ^([0-9]{3})- ]]; then
            # Use base branch prefix instead
            branch_name="$base_branch"
            echo "[specify] Info: Using base branch '$base_branch' to find spec directory" >&2
        else
            # Neither has numeric prefix, fall back to exact match with current branch
            echo "$specs_dir/$branch_name"
            return
        fi
    fi

    # Extract prefix from branch_name (either original or base)
    if [[ "$branch_name" =~ ^([0-9]{3})- ]]; then
        local prefix="${BASH_REMATCH[1]}"
    else
        # Shouldn't happen, but handle gracefully
        echo "$specs_dir/$branch_name"
        return
    fi

    # Search for directories in specs/ that start with this prefix
    local matches=()
    if [[ -d "$specs_dir" ]]; then
        for dir in "$specs_dir"/"$prefix"-*; do
            if [[ -d "$dir" ]]; then
                matches+=("$(basename "$dir")")
            fi
        done
    fi

    # Handle results
    if [[ ${#matches[@]} -eq 0 ]]; then
        # No match found - return the branch name path (will fail later with clear error)
        echo "$specs_dir/$branch_name"
    elif [[ ${#matches[@]} -eq 1 ]]; then
        # Exactly one match - perfect!
        echo "$specs_dir/${matches[0]}"
    else
        # Multiple matches - this shouldn't happen with proper naming convention
        echo "ERROR: Multiple spec directories found with prefix '$prefix': ${matches[*]}" >&2
        echo "Please ensure only one spec directory exists per numeric prefix." >&2
        echo "$specs_dir/$branch_name"  # Return something to avoid breaking the script
    fi
}

get_feature_paths() {
    local repo_root=$(get_repo_root)
    local current_branch=$(get_current_branch)
    local has_git_repo="false"

    if has_git; then
        has_git_repo="true"
    fi

    # Use prefix-based lookup to support multiple branches per spec
    local feature_dir=$(find_feature_dir_by_prefix "$repo_root" "$current_branch")

    cat <<EOF
REPO_ROOT='$repo_root'
CURRENT_BRANCH='$current_branch'
HAS_GIT='$has_git_repo'
FEATURE_DIR='$feature_dir'
FEATURE_SPEC='$feature_dir/spec.md'
IMPL_PLAN='$feature_dir/plan.md'
TASKS='$feature_dir/tasks.md'
RESEARCH='$feature_dir/research.md'
DATA_MODEL='$feature_dir/data-model.md'
QUICKSTART='$feature_dir/quickstart.md'
CONTRACTS_DIR='$feature_dir/contracts'
EOF
}

check_file() { [[ -f "$1" ]] && echo "  ✓ $2" || echo "  ✗ $2"; }
check_dir() { [[ -d "$1" && -n $(ls -A "$1" 2>/dev/null) ]] && echo "  ✓ $2" || echo "  ✗ $2"; }

