# Contribution Guidelines

## Reporting Issues

If you find a bug or have a feature request, please open an issue on GitHub.

## Code Style

- Use `bash` 4.0+ features
- Follow Google's bash style guide
- Use `set -e` for error handling
- Add comments for complex logic
- Use descriptive variable names

## Script Template

```bash
#!/bin/bash
#
# Brief description of what the script does.
#
# Longer description if needed.
#
# Usage: script-name <REPO> <PR_NUMBER> [OPTIONS]
# Example: script-name owner/repo 116
#

set -e

# Arguments
REPO="$1"
PR_NUMBER="$2"

# Validation
if [[ -z "$REPO" ]] || [[ -z "$PR_NUMBER" ]]; then
  echo "Usage: script-name <REPO> <PR_NUMBER> [OPTIONS]" >&2
  echo "Example: script-name owner/repo 116" >&2
  exit 1
fi

# Parse repo
IFS='/' read -r OWNER REPO_NAME <<< "$REPO"
if [[ -z "$REPO_NAME" ]]; then
  echo "Error: REPO must be in format 'owner/name'" >&2
  exit 1
fi

# Script logic
# ...
```

## Testing

Before submitting a PR:

1. Test your script on real PRs
2. Verify error messages are clear
3. Check exit codes are appropriate
4. Ensure JSON output is valid (pipe through `jq`)
5. Test with different repositories

## Documentation

- Update README.md with usage examples
- Document exit codes for tools that use them
- Include real output examples
- Note any required GitHub CLI scopes

## Pull Requests

1. Keep PRs focused on a single feature or fix
2. Write clear commit messages
3. Update documentation as needed
4. Reference any related issues

## Questions?

Open an issue on GitHub or contact the maintainers.
