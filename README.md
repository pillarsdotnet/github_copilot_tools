# GitHub Copilot Review Tools

A comprehensive collection of bash scripts for managing GitHub Copilot reviews and review threads on pull requests.

## Overview

These tools automate common workflows for handling GitHub Copilot code reviews, including:
- Querying open and resolved reviews
- Managing review threads
- Detecting sentinel reviews (no new comments)
- Checking for post-review code changes
- Fetching latest CI/workflow results

All scripts use the GitHub CLI (`gh`) and GraphQL API for reliable, efficient PR management.

## Installation

### Quick Start

1. Clone this repository:
```bash
git clone https://github.com/pillarsdotnet/github-copilot-tools.git
cd github-copilot-tools
```

2. Install scripts to your PATH:
```bash
./install.sh
# or manually:
cp bin/* ~/.local/bin/
chmod +x ~/.local/bin/copilot-* ~/.local/bin/*-review* ~/.local/bin/is-* ~/.local/bin/has-*
```

3. Verify installation:
```bash
which copilot-reviews
```

### Requirements

- `bash` 4.0 or later
- `gh` (GitHub CLI) - [install](https://cli.github.com/)
- `jq` (JSON processor) - [install](https://stedolan.github.io/jq/download/)
- GitHub CLI authentication: `gh auth login`

## Available Scripts

### 1. `copilot-reviews` - List open Copilot reviews

Get all unhidden GitHub Copilot reviews on a PR with their IDs and content.

```bash
copilot-reviews <REPO> <PR_NUMBER>
```

**Examples:**
```bash
copilot-reviews owner/repo 116
copilot-reviews owner/repo 116
```

**Output:** JSON array with review metadata
```json
{
  "id": "PRR_kwDOMImQKc8AAAABJ-oWag",
  "databaseId": 4964619882,
  "body": "## Pull request overview\n..."
}
```

---

### 2. `open-review-threads` - List unresolved review threads

Get all unresolved review threads (inline code comments) on a PR.

```bash
open-review-threads <REPO> <PR_NUMBER>
```

**Examples:**
```bash
open-review-threads owner/repo 116
open-review-threads owner/repo 116
```

**Output:** JSON array with thread details
```json
{
  "id": "PRRT_kwDOMImQKc6aOUFg",
  "path": ".vale.ini",
  "line": null,
  "author": "copilot-pull-request-reviewer",
  "body": "`.vale.ini` references the `proselint` style..."
}
```

---

### 3. `resolve-pr-review-threads` - Resolve all open threads

Resolve all currently unresolved review threads on a PR using GraphQL.

```bash
resolve-pr-review-threads <REPO> <PR_NUMBER>
```

**Examples:**
```bash
resolve-pr-review-threads owner/repo 116
resolve-pr-review-threads owner/repo 116
```

**Output:**
```
════════════════════════════════════════════════════════════════
  Resolving review threads for owner/repo #116
════════════════════════════════════════════════════════════════

🔍 Finding unresolved review threads...
✓ Found 2 unresolved thread(s)

Resolving thread: PRRT_kwDOMImQKc6aOUFg...
  ✓ Resolved

...

════════════════════════════════════════════════════════════════
  Summary
════════════════════════════════════════════════════════════════
Resolved:  2
Failed:    0

✓ All review threads resolved!
```

---

### 4. `hide-pr-review` - Hide a review as resolved

Hide an open review by database ID with an optional custom reason.

```bash
hide-pr-review <REPO> <PR_NUMBER> <REVIEW_DATABASE_ID> [REASON]
```

**Examples:**
```bash
hide-pr-review owner/repo 116 4964619882              # Uses reason "Resolved"
hide-pr-review owner/repo 116 4964619882             # Uses reason "Resolved"
hide-pr-review owner/repo 116 4964619882 "Addressed" # Custom reason
```

**Output:**
```
🔍 Fetching review comment ID for review #4964619882 on owner/repo PR #116...
✓ Found comment ID: PRRC_kwDOMImQKc7i6aSG

🔒 Hiding review with reason 'Resolved'...
✅ Successfully hidden review 4964619882 with reason "resolved"!
```

---

### 5. `latest-copilot-review-id` - Get latest review ID

Find the database ID of the most recent Copilot review on a PR.

```bash
latest-copilot-review-id <REPO> <PR_NUMBER>
```

**Examples:**
```bash
latest-copilot-review-id owner/repo 116
latest-copilot-review-id owner/repo 116
```

**Output:** Just the database ID (suitable for piping)
```
4964619882
```

**Piping example:**
```bash
hide-pr-review owner/repo 116 $(latest-copilot-review-id owner/repo 116) "Resolved"
```

---

### 6. `not-sentinel-review` - Check if review has outstanding comments

Check if a Copilot review is NOT a sentinel (still has outstanding comments).

A sentinel review contains "generated no new comments", indicating Copilot found
no new issues. This script returns 0 if the review is NOT a sentinel (still has
comments that need addressing).

```bash
not-sentinel-review <REPO> <PR_NUMBER> <REVIEW_DATABASE_ID>
```

**Examples:**
```bash
not-sentinel-review owner/repo 116 4964619882
if not-sentinel-review owner/repo 116 $(latest-copilot-review-id owner/repo 116); then
  echo "Review has outstanding comments that need fixing"
fi
```

**Exit codes:**
- `0` = Review is NOT a sentinel (has comments to address)
- `1` = Review IS a sentinel (no new comments found)
- `2` = Review not found or error

---

### 7. `has-post-sentinel-commits` - Check for new commits after review

Check if any commits after the latest Copilot review modified files changed by the PR.

```bash
has-post-sentinel-commits <REPO> <PR_NUMBER>
```

**Examples:**
```bash
has-post-sentinel-commits owner/repo 116
if ! has-post-sentinel-commits owner/repo 116; then
  echo "No new commits after review"
fi
```

**Exit codes:**
- `0` = There are post-review commits changing PR files
- `1` = No such commits or no review exists

---

### 8. `latest-ci-result` - Fetch the latest CI result

Get the most recent CI/workflow run status for a repository or specific PR.

```bash
latest-ci-result <REPO> [PR_NUMBER]
```

**Examples:**
```bash
# Latest run across all branches
latest-ci-result owner/repo

# Latest run for a specific PR
latest-ci-result owner/repo 116
latest-ci-result owner/repo 116
```

**Output:** JSON with workflow run details
```json
{
  "conclusion": "success",
  "databaseId": 32172180710,
  "displayTitle": "Running Copilot Code Review",
  "headBranch": "feature/vincrr/pre-commit",
  "name": "Running Copilot Code Review",
  "status": "completed"
}
```

**Exit codes:**
- `0` = CI passed (success or skipped)
- `1` = CI failed, cancelled, or timed out
- `2` = No CI runs found
- `3` = Invalid arguments

**Usage in scripts:**
```bash
# Check if latest run passed
if latest-ci-result owner/repo 116 > /dev/null; then
  echo "CI passed, safe to merge"
else
  echo "CI failed, needs fixes"
fi

# Get detailed status
latest-ci-result owner/repo 116 | jq '.conclusion'
```

---

### 9. `checkout-pr-branch` - Checkout the branch for a PR

Switch to the local branch corresponding to a given PR.

```bash
checkout-pr-branch <REPO> <PR_NUMBER>
```

**Examples:**
```bash
checkout-pr-branch owner/repo 116
checkout-pr-branch owner/repo 116
```

**Output:**
```
✓ Checked out branch: feature/vincrr/pre-commit
54840f6 fix: provide Vale executable for commit-msg hook and exclude Vale config files
```

**Exit codes:**
- `0` = Successfully checked out branch
- `1` = PR not found or branch checkout failed
- `2` = Already on target branch
- `3` = Invalid arguments
- `4` = Branch doesn't exist locally (can fetch with git fetch)

**Usage examples:**
```bash
# Switch to PR branch and review
checkout-pr-branch owner/repo 116
git log --oneline -10
git diff origin/main

# Switch to PR and check status
checkout-pr-branch owner/repo 116 && git status

# Loop through multiple PRs
for pr in 115 116 117; do
  checkout-pr-branch owner/repo $pr
  echo "PR #$pr:"
  git log -1 --oneline
done
```

---

### 10. `is-rebase-only` - Analyze if changes are from rebasing

Compare files changed between branches to determine if a branch contains
only rebased changes or also has new local modifications.

```bash
is-rebase-only <BASE_BRANCH> <HEAD_BRANCH>
```

**Examples:**
```bash
is-rebase-only origin/main feature/my-feature
is-rebase-only origin/main HEAD
```

**Output:**
```
════════════════════════════════════════════════════════════════════
Rebase Analysis: origin/main → feature/vincrr/pre-commit → HEAD
════════════════════════════════════════════════════════════════════

Files changed in PR (BASE → HEAD):
  Total: 26

Files changed locally (HEAD → current):
  Total: 1

Files changed in BOTH (rebased + locally modified):
  Total: 1

════════════════════════════════════════════════════════════════════

⚠ Mixed: Branch has both rebased changes and new local changes
  1 file(s) were modified and then rebased
```

**Analysis outputs:**
- `✓ Pure rebase` = No local changes beyond remote HEAD (only rebasing occurred)
- `✓ No rebase needed` = All changes are new local commits (no rebasing)
- `⚠ Mixed` = Both rebased changes and new local modifications present

**Use cases:**
```bash
# Check if a branch only has rebase changes
is-rebase-only origin/main feature/branch-name

# Check if current branch matches remote HEAD exactly
is-rebase-only origin/main HEAD

# Use in scripts
if is-rebase-only origin/main HEAD | grep -q "Pure rebase"; then
  echo "Ready to push rebase"
fi
```

---

## Common Workflows

### Workflow 1: Check if PR is ready for merge

```bash
#!/bin/bash

REPO="owner/repo"
PR=116
LATEST=$(latest-copilot-review-id "$REPO" "$PR")

if ! not-sentinel-review "$REPO" "$PR" "$LATEST" && ! has-post-sentinel-commits "$REPO" "$PR"; then
  echo "✓ PR is ready for merge!"
  echo "  - Latest Copilot review is a sentinel (no outstanding comments)"
  echo "  - No new commits after review"
else
  echo "✗ PR needs attention"
  if not-sentinel-review "$REPO" "$PR" "$LATEST"; then
    echo "  - Latest review has outstanding comments to address"
  fi
  if has-post-sentinel-commits "$REPO" "$PR"; then
    echo "  - New commits after review"
  fi
fi
```

### Workflow 2: Resolve all issues and hide review

```bash
#!/bin/bash

REPO="owner/repo"
PR=116

# 1. Fix code issues and commit
git add .
git commit -m "fix: address Copilot review comments"
git push

# 2. Resolve all open threads
resolve-pr-review-threads "$REPO" "$PR"

# 3. Hide the old review as resolved
hide-pr-review "$REPO" "$PR" $(latest-copilot-review-id "$REPO" "$PR") "Resolved"

echo "✓ Ready to request new Copilot review"
```

### Workflow 3: List all Copilot feedback

```bash
#!/bin/bash

REPO="owner/repo"
PR=116

echo "=== Open Copilot Reviews ==="
copilot-reviews "$REPO" "$PR" | jq '.body' | head -20

echo ""
echo "=== Unresolved Review Threads ==="
open-review-threads "$REPO" "$PR" | jq '.[] | "\(.path):\(.line): \(.body)"'
```

---

## Repository Parameter

All scripts require the repository in format `owner/repo-name` as the first parameter.

---

## Configuration

### GitHub CLI Setup

Ensure `gh` is authenticated with proper permissions:

```bash
gh auth login
# Select: GitHub.com
# Select: HTTPS
# Select: Authenticate with your GitHub credentials
```

Required scopes: `repo` (for private repos) or `public_repo` (for public repos)

### Environment Variables

- `GH_REPO`: Override default repo (format: `owner/repo-name`)
- `GH_TOKEN`: GitHub authentication token (set by `gh auth login`)

---

## Script Architecture

Each script follows a consistent pattern:

1. **Argument parsing** - Validate PR number and optional repo
2. **GraphQL query** - Fetch data via GitHub API
3. **JSON filtering** - Process results with `jq`
4. **Output** - Display results or return exit code

### GraphQL Examples

```bash
# Get all reviews for a PR
gh api graphql -f query='query {
  repository(owner: "owner", name: "repo") {
    pullRequest(number: 116) {
      reviews(first: 100) { ... }
    }
  }
}'

# Resolve a review thread
gh api graphql -f threadId="PRRT_..." -f query='
mutation ResolveThread($threadId: ID!) {
  resolveReviewThread(input: {threadId: $threadId}) {
    clientMutationId
  }
}'
```

---

## Troubleshooting

### "gh: command not found"
Install GitHub CLI: https://cli.github.com/

### "jq: command not found"
Install jq: https://stedolan.github.io/jq/download/

### "Error: No Copilot reviews found"
- Check that the PR number is correct
- Verify you have access to the repository
- Ensure the PR has Copilot reviews

### "Authentication failed"
Run: `gh auth login` and complete the authentication flow

### Scripts return no output
Add `set -x` to the script for debug output:
```bash
set -x
copilot-reviews 116
```

---

## Development

### Testing

```bash
# Test on a specific PR
copilot-reviews 116
open-review-threads 116
latest-copilot-review-id 116

# Test with a different repo
copilot-reviews 42 owner/other-repo
```

### Contributing

Contributions welcome! Please ensure:
- Scripts follow bash best practices
- Error handling is robust
- Output is clear and parseable
- Documentation is updated

---

## License

MIT License - See LICENSE file for details

---

## Author

Created by Robert Vincent for automated GitHub Copilot review management.

**Repository:** https://github.com/pillarsdotnet/github-copilot-tools

---

## See Also

- [GitHub CLI Documentation](https://cli.github.com/manual/)
- [GitHub GraphQL API](https://docs.github.com/en/graphql)
- [GitHub Copilot Code Review](https://docs.github.com/en/copilot/code-review)
