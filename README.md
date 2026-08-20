# GitHub Copilot Review Tools

A comprehensive collection of bash scripts for managing GitHub Copilot reviews and review threads on pull requests.

## Overview

These tools automate common workflows for handling GitHub Copilot code reviews, including:
- Querying open and resolved reviews
- Managing review threads
- Detecting sentinel reviews (no new comments)
- Checking for post-review code changes
- Fetching latest CI/workflow results
- Discovering and reasoning about stacked-PR chains
- Categorizing review concerns (including suppressed comments) for faster triage

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
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
copilot-reviews [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
copilot-reviews -r owner/repo 116
copilot-reviews 116  # REPO derived from cwd
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
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
open-review-threads [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
open-review-threads -r owner/repo 116
open-review-threads 116  # REPO derived from cwd
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
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
resolve-pr-review-threads [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
resolve-pr-review-threads -r owner/repo 116
resolve-pr-review-threads 116  # REPO derived from cwd
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
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
hide-pr-review [-r REPO] <PR_NUMBER> <REVIEW_DATABASE_ID> [REASON]
```

**Examples:**
```bash
hide-pr-review -r owner/repo 116 4964619882           # Uses reason "Resolved"
hide-pr-review 116 4964619882                        # REPO derived from cwd
hide-pr-review -r owner/repo 116 4964619882 "Addressed" # Custom reason
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
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
latest-copilot-review-id [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
latest-copilot-review-id -r owner/repo 116
latest-copilot-review-id 116  # REPO derived from cwd
```

**Output:** Just the database ID (suitable for piping)
```
4964619882
```

**Piping example:**
```bash
hide-pr-review -r owner/repo 116 $(latest-copilot-review-id -r owner/repo 116) "Resolved"
```

---

### 6. `review-database-ids` - Get all review database IDs

Extract all database IDs from Copilot reviews on a PR, one per line. Useful for
piping to other commands or looping over reviews. REPO is optional; if
omitted it's derived from the cwd's GitHub remote.

```bash
review-database-ids [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
review-database-ids -r owner/repo 116
review-database-ids 116  # REPO derived from cwd
```

**Output:** One databaseId per line
```
4941808346
4952409221
4952485496
4952676973
4953049700
```

**Piping examples:**
```bash
# Count reviews
review-database-ids -r owner/repo 116 | wc -l

# Loop over review IDs
review-database-ids -r owner/repo 116 | while read id; do
  hide-pr-review -r owner/repo 116 "$id"
done

# Store in bash array
mapfile -t REVIEW_IDS < <(review-database-ids -r owner/repo 116)
for id in "${REVIEW_IDS[@]}"; do
  echo "Review: $id"
done

# Combine with other operations
review-database-ids -r owner/repo 116 | head -1 | xargs -I {} hide-pr-review -r owner/repo 116 {}
```

**Exit codes:**
- `0` = Success (even if no reviews found)
- `1` = Invalid arguments

---

### 7. `not-sentinel-review` - Check if review has outstanding comments

Check if a Copilot review is NOT a sentinel (still has outstanding comments).

A sentinel review contains "generated no new comments", indicating Copilot found
no new issues. This script returns 0 if the review is NOT a sentinel (still has
comments that need addressing). REPO is optional; if omitted it's derived
from the cwd's GitHub remote.

```bash
not-sentinel-review [-r REPO] <PR_NUMBER> <REVIEW_DATABASE_ID>
```

**Examples:**
```bash
not-sentinel-review -r owner/repo 116 4964619882
if not-sentinel-review -r owner/repo 116 $(latest-copilot-review-id -r owner/repo 116); then
  echo "Review has outstanding comments that need fixing"
fi
```

**Exit codes:**
- `0` = Review is NOT a sentinel (has comments to address)
- `1` = Review IS a sentinel (no new comments found)
- `2` = Review not found or error

---

### 8. `has-post-sentinel-commits` - Check for new commits after review

Check if any commits after the latest Copilot review modified files changed by the PR.
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
has-post-sentinel-commits [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
has-post-sentinel-commits -r owner/repo 116
if ! has-post-sentinel-commits -r owner/repo 116; then
  echo "No new commits after review"
fi
```

**Exit codes:**
- `0` = There are post-review commits changing PR files
- `1` = No such commits or no review exists

---

### 9. `latest-ci-result` - Fetch the latest CI result

Get the most recent CI/workflow run status for a repository or specific PR.
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.
Optionally pin the result to a specific workflow via `WORKFLOW` - if a repo
runs more than one workflow on the same branch (e.g. a "Code checker" CI
workflow alongside an unrelated "Running Copilot Code Review" workflow),
"most recent run" can silently mean the wrong one and report a false
positive/negative. Pass `WORKFLOW` (and `""` for PR_NUMBER if omitting it)
whenever the repo has more than one workflow that could run on a branch.

```bash
latest-ci-result [-r REPO] [PR_NUMBER] [WORKFLOW]
```

**Examples:**
```bash
# Latest run across all branches
latest-ci-result -r owner/repo

# Latest run for a specific PR
latest-ci-result -r owner/repo 116

# REPO derived from cwd's GitHub remote
latest-ci-result 116

# Latest run of a specific workflow for a specific PR
latest-ci-result -r owner/repo 116 "Code checker"

# Latest run of a specific workflow, no PR filter
latest-ci-result -r owner/repo "" "Code checker"
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
if latest-ci-result -r owner/repo 116 > /dev/null; then
  echo "CI passed, safe to merge"
else
  echo "CI failed, needs fixes"
fi

# Get detailed status
latest-ci-result -r owner/repo 116 | jq '.conclusion'
```

---

### 10. `checkout-pr-branch` - Checkout the branch for a PR

Switch to the local branch corresponding to a given PR.
-r/REPO is optional; if omitted it's derived from the cwd's GitHub remote via `gh repo view`.

```bash
checkout-pr-branch [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
checkout-pr-branch -r owner/repo 116
checkout-pr-branch 116  # REPO derived from cwd
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
checkout-pr-branch -r owner/repo 116
git log --oneline -10
git diff origin/main

# Switch to PR and check status
checkout-pr-branch -r owner/repo 116 && git status

# Loop through multiple PRs
for pr in 115 116 117; do
  checkout-pr-branch -r owner/repo $pr
  echo "PR #$pr:"
  git log -1 --oneline
done
```

---

### 11. `check-changed-files` - Count the number of files changed beyond a simple rebase

Compare files changed between branches to determine if a branch contains
only rebased changes or also has new local modifications.

```bash
check-changed-files <BASE_BRANCH> <HEAD_BRANCH>
```

**Examples:**
```bash
check-changed-files origin/main feature/my-feature
check-changed-files origin/main HEAD
```

**Output:**
```
════════════════════════════════════════════════════════════════════════
Rebase Analysis: origin/main → feature/vincrr/pre-commit → HEAD
════════════════════════════════════════════════════════════════════════

Files changed in PR (BASE → HEAD):
  Total: 26

Files changed locally (HEAD → current):
  Total: 1

Files changed in BOTH (rebased + locally modified):
  Total: 1

════════════════════════════════════════════════════════════════════════

⚠ Mixed: Branch has both rebased changes and new local changes
  1 file(s) were modified and then rebased
```

**Analysis outputs:**
- `✓ Pure rebase` = No local changes beyond remote HEAD (only rebasing occurred)
- `✓ No rebase needed` = All changes are new local commits (no rebasing)
- `⚠ Mixed` = Both rebased changes and new local modifications present

**Exit codes:**
- `-1` = Error in arguments or processing
- `0+` = Number of files changed beyond a simple rebase

**Use cases:**
```bash
# Check if a branch only has rebase changes
CHANGED_FILES=$(check-changed-files origin/main feature/branch-name)
if [ "$CHANGED_FILES" -eq 0 ]; then
  echo 'Rebase only; no local changes'
fi

# Use in scripts
if check-changed-files origin/main HEAD > /dev/null; then
  echo "Files changed: $(check-changed-files origin/main HEAD)"
fi
```

### 12. `refresh-needed` - Determine if a new review should be requested

Check if a new Copilot review is needed based on whether all previous reviews are
resolved and new files have been changed.

```bash
refresh-needed [-r REPO] <PR_NUMBER> <FILES_CHANGED_COUNT>
```

**Arguments:**
- `-r REPO` - Repository in format `owner/repo-name` (optional; if omitted
  it's derived from the cwd's GitHub remote)
- `PR_NUMBER` - Pull request number
- `FILES_CHANGED_COUNT` - Number of files changed since last review (typically from `check-changed-files`)

**Examples:**
```bash
refresh-needed -r owner/repo 116 3
refresh-needed 116 3  # REPO derived from cwd
```

**Output:**
```
✓ New review needed: All previous reviews resolved, 3 file(s) changed
```

or

```
✗ New review not needed:
  - Outstanding comments remain in previous reviews
  - No files changed since last review
```

**Exit codes:**
- `0` = New review IS needed (all reviews resolved AND files changed)
- `1` = New review NOT needed (no changes or reviews still outstanding)

**Use cases:**
```bash
# Check if review refresh is needed
CHANGED=$(check-changed-files origin/main HEAD)
if refresh-needed -r owner/repo 116 "$CHANGED"; then
  echo "✓ Ready to request new Copilot review"
  request-pr-review -r owner/repo 116
else
  echo "✗ Address outstanding comments before requesting new review"
fi

# Combine with other workflow steps
if refresh-needed -r "$REPO" "$PR" 0; then
  echo "No files changed but all reviews are resolved"
fi
```

---

### 13. `analyze-pr-stack` - Discover the full stacked-PR chain

Given any PR number in a stack, walks down through base branches to find the
root (base = default branch), then walks up through head branches to find
every PR chained on top of it. Handles GitHub's automatic base-retargeting
after a PR merges (which can otherwise obscure the chain), and reports which
PRs are still open vs. already closed/merged. REPO is optional; if omitted
it's derived from the cwd's GitHub remote.

```bash
analyze-pr-stack [-r REPO] <PR_NUMBER> [--json]
```

**Examples:**
```bash
analyze-pr-stack -r owner/repo 110
analyze-pr-stack 110 --json  # REPO derived from cwd
```

**Output (text):** Indented tree from the default branch to the top PR,
marking closed/merged PRs as `(state, skip)`.

**Output (--json):**
```json
{
  "stack": [115, 120],
  "order": "bottom-to-top (root -> leaves)",
  "root_branch": "main",
  "leaves": [115, 120],
  "closed": [108, 109, 110, 111, 116, 117, 112]
}
```

- `stack` - currently open PRs, in bottom-to-top order (the ones that still need processing)
- `closed` - PRs found while tracing the chain that are already closed or merged

**Exit codes:**
- `0` = Stack resolved successfully
- `1` = Invalid arguments
- `2` = Starting PR not found, or repo/default branch could not be determined

**Usage in scripts:**
```bash
# Feed the live open stack straight into stack-cycle-next-pr
analyze-pr-stack -r owner/repo 110 --json | jq -r '.stack[]' | stack-cycle-next-pr -r owner/repo --json
```

---

### 14. `categorize-pr-reviews` - Categorize suppressed and open review concerns

Combines suppressed comments (parsed out of Copilot review bodies) with open
review threads, and tags each with a best-guess concern category (bug,
style, documentation, performance, scope, testing) by keyword matching.
Surfaces suppressed comments prominently since Copilot hides them by default
but they still represent real concerns that must be remediated. REPO is
optional; if omitted it's derived from the cwd's GitHub remote.

```bash
categorize-pr-reviews [-r REPO] <PR_NUMBER> [--json]
```

**Examples:**
```bash
categorize-pr-reviews -r owner/repo 117
categorize-pr-reviews 117 --json  # REPO derived from cwd
```

**Output (--json):**
```json
{
  "suppressed": [
    {"file": "README.md", "line": 61, "text": "...", "category": "bug"}
  ],
  "non_suppressed": [
    {"file": "Y.ps1", "line": 15, "text": "...", "category": "style", "thread_id": "PRRT_..."}
  ],
  "summary": {
    "total": 4,
    "suppressed": 4,
    "non_suppressed": 0,
    "by_category": {"bug": 2, "style": 2}
  }
}
```

**Exit codes:**
- `0` = Success (even if no concerns found)
- `1` = Invalid arguments

**Note:** Categories are heuristic keyword-match hints meant to speed up
triage, not a definitive classification—always read `text` before deciding
how to remediate.

---

### 15. `stack-cycle-next-pr` - Determine which PRs in a stack still need work

Given an ordered PR stack, checks each PR's latest Copilot review with
`is-sentinel-review` and `has-post-sentinel-commits` to decide whether it's
already stable (skip) or needs another pass (process). Automates the manual
"cycle back through the stack" bookkeeping. REPO is optional; if omitted
it's derived from the cwd's GitHub remote.

```bash
stack-cycle-next-pr [-r REPO] <PR_NUMBER...> [--json]
```

PR numbers may also be piped on stdin, one per line—handy for chaining
directly off `analyze-pr-stack`:

```bash
analyze-pr-stack -r owner/repo 110 --json | jq -r '.stack[]' | stack-cycle-next-pr -r owner/repo --json
```

**Examples:**
```bash
stack-cycle-next-pr -r owner/repo 115 116 117 108 109 110 111 120
stack-cycle-next-pr -r owner/repo 115 116 117 108 109 110 111 120 --json
```

**Output (--json):**
```json
{
  "next_pr": 110,
  "skip": [116, 117, 108, 109],
  "process": [110, 111, 120],
  "all_stable": false,
  "progress": "4 of 8 PRs stable"
}
```

**Exit codes:**
- `0` = More PRs to process
- `1` = All PRs stable (workflow complete)
- `2` = Invalid arguments

---

### 16. `apply-pr-diffs` - Apply a PR's changes as per-file diffs

Rebuilds a PR's changes onto a fresh branch by diffing each file between
two refs and applying those diffs directly, instead of replaying commit
history. Useful when cherry-pick or rebase produce excessive conflicts
because commits are tangled or touch overlapping config files. Only
prepares and stages the result locally—does not commit or push.

```bash
apply-pr-diffs <BASE_REF> <PR_REF> [CLEAN_BRANCH] [--force]
```

**Examples:**
```bash
apply-pr-diffs origin/main origin/feature/x
apply-pr-diffs origin/main origin/feature/x feature/x-clean --force
```

**Exit codes:**
- `0` = All diffs applied cleanly (pre-commit clean or skipped)
- `1` = One or more diffs failed to apply (conflict)—diffs are preserved for inspection
- `2` = All diffs applied, but pre-commit found issues
- `3` = Invalid arguments, not a git repo, dirty working tree, or branch already exists (pass `--force`)

---

### 17. `auto-rebase-pr` - Rebase a PR with automatic fallback

Orchestrates the full STEP 2 rebase decision: tries a direct rebase first,
falls back to a squashed cherry-pick, then falls back to `apply-pr-diffs`
as a last resort. **Designed to fail loudly if anything goes wrong**, so you know
immediately when manual intervention is needed rather than discovering issues
much later. **Does not push by default**—force-pushing a shared branch is hard
to reverse, so pass `--push` explicitly, or run the printed command yourself
after reviewing the result.

**IMPORTANT:** Never force-push if the script exits with code 3 or 6.
Instead, inspect the diagnostic output and decide whether to (a) rebase manually,
(b) delete the local branches and try again, or (c) mark the PR as requiring
special handling. REPO is optional; if omitted it's derived from the cwd's
GitHub remote.

```bash
auto-rebase-pr [-r REPO] <PR_NUMBER> [--push] [--force]
```

**Examples:**
```bash
auto-rebase-pr -r owner/repo 116          # prepare locally, print the push command
auto-rebase-pr -r owner/repo 116 --push   # prepare and force-push the result
auto-rebase-pr 116 --push                 # REPO derived from cwd
```

**Exit codes:**
- `0` = Succeeded via direct rebase (clean)
- `1` = Succeeded via cherry-pick (squashed into one commit)
- `2` = Succeeded via per-file diff application (`apply-pr-diffs`)
- `3` = All strategies failed; do NOT push, inspect diagnostic output
- `4` = Invalid arguments or setup error (fix parameters and retry)
- `5` = Local result prepared successfully, but `--push` failed; retry with `--push`
- `6` = Strategy partially succeeded but left working tree dirty; manual cleanup needed

**Handling failure cases:**
- Exit code `3` (all strategies failed): One or more commits have conflicts that
  cannot be resolved automatically. You will need to rebase manually to understand
  the conflicts, or mark the PR as requiring special handling.
- Exit code `6` (dirty working tree): A rebase strategy partially succeeded but
  left uncommitted changes. This usually indicates a pre-commit hook or other
  tool modified files unexpectedly. Run `git status` to see what happened, then
  either manually fix and commit, or delete branches and start over.

**Note:** When strategy 2 or 3 succeeds, the commit message is the PR title
verbatim—review/amend it to conform to conventional-commits before pushing.

---

### 18. `reset-copilot-reviewer` - Force reset Copilot reviewer

Forcibly remove GitHub Copilot from the PR's reviewers, then immediately re-add it.
This works around issues where the standard "request review" button in the GitHub UI
fails silently or doesn't trigger a new review. By removing and re-adding Copilot,
you force GitHub to send a fresh review request. REPO is optional; if
omitted it's derived from the cwd's GitHub remote.

```bash
reset-copilot-reviewer [-r REPO] <PR_NUMBER>
```

**Examples:**
```bash
reset-copilot-reviewer -r owner/repo 116
reset-copilot-reviewer 116  # REPO derived from cwd
```

**Output:**
```
════════════════════════════════════════════════════════════════
  Reset Copilot Reviewer on PR #116
════════════════════════════════════════════════════════════════

Step 1: Removing copilot-pull-request-reviewer from reviewers...
✓ Removed copilot-pull-request-reviewer

Step 2: Re-adding copilot-pull-request-reviewer as reviewer...
✓ Re-added copilot-pull-request-reviewer

════════════════════════════════════════════════════════════════
✓ Success! Copilot review request has been reset.
  GitHub should now send a fresh review request.
════════════════════════════════════════════════════════════════
```

**Exit codes:**
- `0` = Success (Copilot removed and re-added)
- `1` = Error (invalid repo format, PR not found, gh command failed, etc.)

**When to use:**
- Copilot review request button in GitHub UI doesn't respond
- Previous review request appears to have been lost or ignored
- Need to force a fresh code review without changing the PR
- Part of the PR review workflow when step 11 (request review) needs a workaround

**Troubleshooting:**
- If Copilot wasn't previously a reviewer, the remove step will warn but continue
- Ensure `gh` is authenticated: `gh auth login`
- Verify the PR number is correct: `gh pr view <REPO> <PR_NUMBER>`

---

### 19. `analyze-ci-failures` - Classify what failed in the latest CI run

Fetches the latest CI run for a PR and classifies failures instead of making
the AI read raw workflow logs. Combines `latest-ci-result` (run resolution),
`gh run view --json jobs` (accurate failed job/step names-raw log step names
are frequently "UNKNOWN STEP" due to a `gh` limitation), and `--log-failed`
log content (pre-commit hook id/exit code extraction, generic error-pattern
matching for test/git/external-dependency failures).

`WORKFLOW` is forwarded straight to `latest-ci-result` to pin the run to a
specific workflow - pass it whenever the repo has more than one workflow
that could run on a branch (see `latest-ci-result` above for why). REPO is
optional; if omitted it's derived from the cwd's GitHub remote.

```bash
analyze-ci-failures [-r REPO] <PR_NUMBER> [WORKFLOW] [--json]
```

**Examples:**
```bash
analyze-ci-failures -r mge0/DevOps 149
analyze-ci-failures -r mge0/DevOps 149 --json
analyze-ci-failures -r mge0/DevOps 149 "Code checker"
analyze-ci-failures 149 "Code checker"  # REPO derived from cwd
analyze-ci-failures -r mge0/DevOps 149 "Code checker" --json
```

**Output (--json):**
```json
{
  "conclusion": "failure",
  "run_url": "https://github.com/mge0/DevOps/actions/runs/32361852145",
  "failed_steps": [{"job": "Lint changed files", "step": "Sync Vale downloaded packages"}],
  "failures": [
    {
      "type": "pre_commit_hook",
      "hook_id": "vale-sync",
      "occurrences": 3,
      "requires_ai": true,
      "reason": "not a known auto-fixable formatter hook; needs investigation"
    }
  ],
  "summary": {"total": 1, "auto_fixable": 0, "requires_ai": 1}
}
```

Repeated retries of the same hook (a retry-loop step, as with `vale-sync`
above) collapse into one entry with an `occurrences` count rather than being
reported as separate failures.

**Exit codes:**
- `0` = CI passed; nothing to analyze
- `1` = CI failed; failures were parsed and reported
- `2` = No CI run found, or the run is still in progress
- `3` = Invalid arguments

**Note:** `requires_ai` is `false` only for a small known set of auto-fixable
formatter hooks (see `fix-precommit-failures` below)-everything else defaults
to `true`, deliberately, so nothing gets silently mis-classified as handled.

---

### 20. `fix-precommit-failures` - Auto-converge pre-commit hooks locally

Runs `pre-commit run --all-files` repeatedly (up to a pass limit) and lets
its own fixers converge-hooks like `trailing-whitespace`, `end-of-file-fixer`,
and `prettier` modify files in place and exit nonzero just to signal "I
changed something," so rerunning after that succeeds. Reports whichever
hooks are *still* failing after convergence (nothing left to auto-fix) for
AI/human attention, rather than guessing at a fix. Deliberately not a
per-hook-name command lookup table-hand-mapping "hook X → fixer command Y"
is brittle and silently wrong for hooks whose failure isn't a formatting
issue at all (e.g. a misconfigured hook, or `gitleaks`).

```bash
fix-precommit-failures [MAX_PASSES]
```

Must be run from inside a git working tree with a `.pre-commit-config.yaml`.
Leaves auto-fixed changes staged (`git add`) but does not commit.

**Examples:**
```bash
fix-precommit-failures        # up to 3 passes (default)
fix-precommit-failures 5      # up to 5 passes
```

**Exit codes:**
- `0` = All hooks pass after convergence
- `1` = Some hooks still fail after convergence; see the reported list
- `2` = pre-commit not available, or no `.pre-commit-config.yaml` here
- `3` = Not inside a git repository

---

## Workflow Examples

### Example 1: Check if review refresh is needed

```bash
#!/bin/bash

REPO="owner/repo"
PR=116

# Get number of files changed since last review
CHANGED_FILES=$(check-changed-files origin/main HEAD)

# Check if new review should be requested
if refresh-needed -r "$REPO" "$PR" "$CHANGED_FILES"; then
  echo "✓ All reviews resolved and new files changed"
  echo "  Ready to request new Copilot review"
else
  echo "✗ Not ready for new review yet"
fi
```

### Example 2: List all Copilot feedback

```bash
#!/bin/bash

REPO="owner/repo"
PR=116

echo "=== Open Copilot Reviews ==="
copilot-reviews -r "$REPO" "$PR" | jq '.body' | head -20

echo ""
echo "=== Unresolved Review Threads ==="
open-review-threads -r "$REPO" "$PR" | jq '.[] | "\(.path):\(.line): \(.body)"'
```

---

## Repository Parameter

Every script that talks to a specific PR accepts an optional `-r REPO` flag
(`owner/repo-name` format), which may appear before or after the other
arguments. When `-r` is omitted, REPO is derived from the current
directory's GitHub remote via `gh repo view` - so scripts run from inside
a repo checkout usually don't need `-r` at all.

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

# Test with a different repo (-r may appear anywhere in argv)
copilot-reviews 42 -r owner/other-repo
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
