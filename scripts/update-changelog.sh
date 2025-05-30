#!/bin/bash

# Script to help update the CHANGELOG.md file
# Usage: ./scripts/update-changelog.sh [type] [description]
# Types: breaking, feature, bugfix, security, deprecated, removed

set -e

CHANGELOG_FILE="CHANGELOG.md"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CHANGELOG_PATH="$PROJECT_ROOT/$CHANGELOG_FILE"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_usage() {
    echo -e "${BLUE}Usage: $0 [type] [description]${NC}"
    echo ""
    echo -e "${YELLOW}Types:${NC}"
    echo "  breaking   - Breaking changes that require user action"
    echo "  feature    - New functionality and enhancements"
    echo "  bugfix     - Bug fixes and corrections"
    echo "  security   - Security-related changes"
    echo "  deprecated - Features that will be removed in future versions"
    echo "  removed    - Features that have been removed"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 feature \"Add support for custom metrics endpoint\""
    echo "  $0 bugfix \"Fix memory leak in DNS exporter (#123)\""
    echo "  $0 breaking \"Remove deprecated --old-flag parameter\""
}

if [ $# -eq 0 ]; then
    print_usage
    exit 1
fi

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    print_usage
    exit 0
fi

if [ $# -ne 2 ]; then
    echo -e "${RED}Error: Both type and description are required${NC}"
    print_usage
    exit 1
fi

TYPE="$1"
DESCRIPTION="$2"

# Validate type
case "$TYPE" in
    breaking|feature|bugfix|security|deprecated|removed)
        ;;
    *)
        echo -e "${RED}Error: Invalid type '$TYPE'${NC}"
        print_usage
        exit 1
        ;;
esac

# Check if changelog file exists
if [ ! -f "$CHANGELOG_PATH" ]; then
    echo -e "${RED}Error: CHANGELOG.md not found at $CHANGELOG_PATH${NC}"
    exit 1
fi

# Map type to section header
case "$TYPE" in
    breaking)
        SECTION="BREAKING CHANGES"
        ;;
    feature)
        SECTION="Features"
        ;;
    bugfix)
        SECTION="Bugfix"
        ;;
    security)
        SECTION="Security"
        ;;
    deprecated)
        SECTION="Deprecated"
        ;;
    removed)
        SECTION="Removed"
        ;;
esac

# Create a temporary file for processing
TEMP_FILE=$(mktemp)

# Process the changelog with a simpler, more robust approach
{
    # Read and copy everything until we find the [Unreleased] section
    while IFS= read -r line; do
        echo "$line"
        if [[ "$line" == "## [Unreleased]" ]]; then
            break
        fi
    done
    
    # Now we're after the [Unreleased] line
    # Look for existing content under Unreleased section
    found_section=false
    added_entry=false
    in_unreleased=true
    buffer=""
    
    while IFS= read -r line && [ "$in_unreleased" = true ]; do
        # Check if we've reached the next version section
        if [[ "$line" =~ ^##\ \[.*\]\ -\ [0-9] ]]; then
            # We've reached the first versioned release
            # If we haven't added our entry yet, add it now
            if [ "$added_entry" = false ]; then
                if [ "$found_section" = false ]; then
                    echo ""
                    echo "### $SECTION"
                    echo "- $DESCRIPTION"
                else
                    # We found the section but haven't added our entry yet
                    # Add it to the buffer and then flush
                    echo "- $DESCRIPTION"
                fi
                echo ""
            fi
            echo "$line"
            in_unreleased=false
            break
        fi
        
        # Check if this is our target section
        if [[ "$line" == "### $SECTION" ]]; then
            echo "$line"
            found_section=true
            # Read the next line to see if we should add our entry immediately
            continue
        fi
        
        # Check if this is a different section
        if [[ "$line" =~ ^###\  ]] && [[ "$line" != "### $SECTION" ]]; then
            # Different section found
            if [ "$found_section" = false ]; then
                # Add our section before this one
                echo ""
                echo "### $SECTION"
                echo "- $DESCRIPTION"
                echo ""
                added_entry=true
            fi
            echo "$line"
            continue
        fi
        
        # If we're in our target section, add our entry before any existing entries
        if [ "$found_section" = true ] && [ "$added_entry" = false ]; then
            echo "- $DESCRIPTION"
            added_entry=true
        fi
        
        echo "$line"
    done
    
    # If we never found any sections in unreleased, add our section now
    if [ "$in_unreleased" = true ] && [ "$added_entry" = false ]; then
        echo ""
        echo "### $SECTION"
        echo "- $DESCRIPTION"
        echo ""
    fi
    
    # Copy the rest of the file
    while IFS= read -r line; do
        echo "$line"
    done
} < "$CHANGELOG_PATH" > "$TEMP_FILE"

# Replace the original file
mv "$TEMP_FILE" "$CHANGELOG_PATH"

echo -e "${GREEN}Successfully added entry to CHANGELOG.md:${NC}"
echo -e "${YELLOW}Section:${NC} $SECTION"
echo -e "${YELLOW}Description:${NC} $DESCRIPTION"
echo ""
echo -e "${BLUE}Tip:${NC} Review the changes with: git diff CHANGELOG.md" 