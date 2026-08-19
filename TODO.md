# Tooling Opportunities to Reduce AI Credits Usage

## Overview

This document identifies workflow steps in the stacked PR review process that would benefit from additional tooling. Each opportunity represents a place where AI is currently "thinking about" a problem that could be solved by a prepared, deterministic script instead—thereby reducing GitHub Copilot credits consumption.

## Identified Opportunities

### PRIORITY 1: Highest Impact (Recurring, AI-Intensive)

#### 1.1 Auto-Rebase with Intelligent Fallback (`auto-rebase-pr`)

**Current State**: Manual decision logic between cherry-pick and per-file diff approaches.

**Workflow Step**: STEP 2 (Rebase on base branch)

**Problem**: 
- AI evaluates rebase conflicts and decides cherry-pick vs. per-file diff
- This requires analyzing commit structure, dependency chains, and conflict severity
- Decision is made fresh for each PR despite consistent patterns

**Solution**: Create `auto-rebase-pr` script that:
1. Attempts `git rebase origin/<BASE>` and captures output
2. If rebase succeeds → push and exit
3. If conflicts detected → attempt cherry-pick (commits from PR branch)
4. If cherry-pick also fails → fallback to per-file diff approach
5. For per-file diff: automatically create diffs, apply them, verify against PR description
6. Return exit code indicating which strategy succeeded

**Impact**: 
- Eliminates AI analysis of "which rebase strategy to use"
- Automates entire rebase-to-push workflow for ~70% of PRs
- Reduces AI tokens by ~500-1000 per PR

**Effort**: Medium (bash/Go, requires testing on 5-10 real PRs)

---

#### 1.2 Review Comment Categorizer & Summarizer (`categorize-pr-reviews`)

**Current State**: Manual review of all comments, categorization by type (suppressed vs. non-suppressed, then by concern category).

**Workflow Step**: STEP 5-6 (Collect review comments and open threads)

**Problem**:
- AI reads entire review output and categorizes by hand
- Must distinguish suppressed comments (CRITICAL) from non-suppressed
- Must infer concern category (bug, style, doc, performance, scope, etc.)
- Process repeats for every PR

**Solution**: Create `categorize-pr-reviews` script that:
1. Fetch all reviews via `copilot-reviews`
2. Extract suppressed comments (from `<details>` HTML tags)
3. Categorize each comment by pattern matching:
   - `null`, `error`, `exception`, `fail`, `undefined` → **Bug**
   - `consider`, `suggest`, `could`, `style` → **Style**
   - `typo`, `comment`, `docstring`, `README` → **Documentation**
   - `performance`, `optimize`, `complexity` → **Performance**
   - `scope`, `PR description`, `unexpected` → **Scope**
   - `test`, `verify`, `coverage` → **Testing**
4. Output structured JSON with:
   ```json
   {
     "suppressed": [
       {"file": "X.ps1", "line": 42, "category": "bug", "text": "..."}
     ],
     "non_suppressed": [
       {"file": "Y.ps1", "line": 15, "category": "style", "text": "..."}
     ],
     "summary": {
       "total": 12,
       "suppressed": 4,
       "by_category": {"bug": 2, "style": 5, "doc": 3, ...}
     }
   }
   ```
5. Print structured report for AI to action

**Impact**:
- Eliminates AI manual categorization work
- Highlights suppressed comments automatically (ensures they are not skipped)
- Provides structured input for remediation (faster AI reasoning)
- Reduces AI tokens by ~300-500 per PR

**Effort**: Medium (bash/Go + pattern database)

---

#### 1.3 Cycle-Back Stack Orchestrator (`stack-cycle-next-pr`)

**Current State**: Manual iteration through PR stack, manually checking each PR for sentinel review + no new commits.

**Workflow Step**: STEP 13+ (Cycle back through stack)

**Problem**:
- After processing last PR, AI must manually check each PR's review state
- AI iterates through 5-9 PRs, checking review status and commit history
- Decision logic for "skip this PR?" is done per PR
- This repeats every cycle through the stack

**Solution**: Create `stack-cycle-next-pr` script that:
1. Accept PR stack as ordered list (e.g., `116 117 108 109 110 111 112 115 120`)
2. For each PR in order:
   - Check if it has a sentinel review (from `is-sentinel-review`)
   - Check if it has new commits since that review (via `has-post-sentinel-commits`)
   - If both true → **mark as SKIP**
   - If false for either → **mark as PROCESS**
3. Output next PR to process and skip list:
   ```json
   {
     "next_pr": 110,
     "skip": [116, 117, 108, 109],
     "process": [110, 111, 112, 115, 120],
     "all_stable": false,
     "progress": "4 of 9 PRs stable"
   }
   ```
4. Exit code: 0 = more PRs to process, 1 = all stable (workflow complete)

**Impact**:
- Eliminates entire manual cycle-back decision logic
- Orchestrates multi-PR workflows autonomously
- Enables batch processing (AI can say "process next 3 PRs" and script tells it which)
- Reduces AI tokens by ~200-300 per cycle
- Enables running entire stack workflow with AI only for remediation steps

**Effort**: Medium (Go/bash, extends existing `is-sentinel-review` and `has-post-sentinel-commits`)

---

### PRIORITY 2: Medium Impact (Recurring, Moderate AI Involvement)

#### 2.1 Per-File Diff Automation (`apply-pr-diffs`)

**Current State**: Manual creation of per-file diffs, manual application and verification.

**Workflow Step**: STEP 2 (Rebase using per-file diff approach)

**Problem**:
- AI creates diffs manually for each file
- AI applies diffs one by one
- AI verifies result matches PR description

**Solution**: Create `apply-pr-diffs` script that:
1. Accept two refs: `BASE` and `PR_BRANCH`
2. Identify all files changed: `git diff --name-only $BASE..$PR_BRANCH`
3. Create diffs for each file
4. Create fresh branch: `git checkout -b <PR>-clean origin/$BASE`
5. Apply each diff: `git apply <diff>`
6. On conflict → output which diff failed (let AI decide how to fix)
7. After all diffs applied:
   - Run pre-commit to catch obvious issues
   - List all files staged
   - Return exit code: 0 = success, 1 = conflicts, 2 = pre-commit failed
8. Print summary with file count, conflict details

**Impact**:
- Eliminates manual per-file diff workflow (currently ~400 tokens per PR)
- Automates cleanup and formatting
- Structured conflict reporting for AI decision-making
- Reduces AI tokens by ~200-400 per complex rebase

**Effort**: Medium (bash, wraps git commands)

---

#### 2.2 Review Comment Remediation Guide (`remediation-strategies`)

**Current State**: AI reads each comment and decides unilaterally how to remediate.

**Workflow Step**: STEP 7 (Remediate all concerns)

**Problem**:
- AI analyzes each comment independently
- Decision of "fix code" vs. "update docs" vs. "add comment" requires context
- No pre-computed strategies for common concern patterns

**Solution**: Create `remediation-strategies` script that:
1. Accept structured review comment (from `categorize-pr-reviews`)
2. Look up concern category + keywords + file type
3. Output suggested remediation strategy:
   ```json
   {
     "comment": "Consider adding null-safe error handling",
     "category": "bug",
     "file_type": "powershell",
     "strategies": [
       {
         "type": "code_fix",
         "description": "Add null check before accessing property",
         "example": "if ($var -ne $null) { ... }"
       },
       {
         "type": "comment",
         "description": "Justify current approach in code comment",
         "example": "# Null check not needed here because ..."
       },
       {
         "type": "test",
         "description": "Add unit test for null input case",
         "example": "Test-Invoke-Func -InputNull"
       }
     ],
     "recommended": "code_fix"
   }
   ```
4. Built-in strategy database for:
   - Logic/null-safety issues → "add null check"
   - Performance issues → "optimize algorithm" or "add caching"
   - Documentation issues → "update comment/docstring"
   - Style issues → "run prettier/format" or "apply linter fix"
   - Scope issues → "update PR description" or "move to different PR"

**Impact**:
- Converts AI's unilateral decision into a guided selection
- Provides pre-vetted strategies for common patterns
- Reduces AI reasoning per-comment by ~50 tokens
- Improves consistency across PR reviews
- Reduces AI tokens by ~100-200 per PR

**Effort**: Medium (bash/Go + strategy database)

---

#### 2.3 CI Failure Analyzer (`analyze-ci-failures`)

**Current State**: AI reads CI logs, identifies failure type, suggests fixes.

**Workflow Step**: STEP 4 (Remediate CI failures)

**Problem**:
- AI must parse logs to find error messages
- AI infers the cause (pre-commit, linter, test, etc.)
- AI suggests fix for each failure

**Solution**: Create `analyze-ci-failures` script that:
1. Fetch latest run via `latest-ci-result`
2. Parse logs for error patterns:
   - `vale`, `prettier`, `PSScriptAnalyzer` → pre-commit hook failures
   - `Test failed`, `assert`, `error:` → test/logic failures
   - `Kusto.Language`, `BlobNotFound` → external dependency failures (skip)
   - `git`, `rebase`, `merge` → git operation failures
3. Output structured report:
   ```json
   {
     "failures": [
       {
         "type": "pre-commit_hook",
         "hook": "prettier",
         "files": ["Runbooks/AdoServiceHookMonitor-Config.json"],
         "fix_command": "npx prettier --write Runbooks/AdoServiceHookMonitor-Config.json",
         "requires_ai": false
       },
       {
         "type": "logic_failure",
         "hook": "test",
         "test_name": "TestNullInput",
         "output": "Expected null, got error",
         "requires_ai": true
       }
     ],
     "action": "Run fix_commands for non-AI failures, then request AI for 1 logic failure"
   }
   ```

**Impact**:
- Automates formatting/linting fixes (no AI needed)
- Categorizes failures for AI to focus on real issues
- Saves ~100-300 tokens per PR with pre-commit-only failures
- Reduces AI tokens by ~200-500 per PR with CI failures

**Effort**: Medium (bash/Go + error pattern database)

---

### PRIORITY 3: Lower Impact but High Convenience

#### 3.1 PR Stack Structure Analyzer (`analyze-pr-stack`)

**Current State**: AI manually understands PR dependency chain from PR descriptions/branch names.

**Workflow Step**: Informational (used at beginning to understand stack order)

**Problem**:
- No single source of truth for PR stack structure
- AI infers from branch naming and PR descriptions
- Risk of processing PRs in wrong order

**Solution**: Create `analyze-pr-stack` script that:
1. Accept repo and starting PR number
2. Traverse dependency chain:
   - For each PR: get `baseRefName` via `gh pr view`
   - Follow base to next PR on same base
   - Build graph: PR → base → next PR up the chain
3. Output stack as tree:
   ```
   main
   └── feature/vincrr/pr-template (PR #115)
       └── feature/vincrr/pre-commit (PR #116)
           └── feature/vincrr/ci-code-checker (PR #117)
               └── feature/vincrr/phase-1 (PR #108)
                   └── feature/vincrr/phase-2 (PR #109)
                       └── ...
   ```
4. Output as JSON for scripting:
   ```json
   {
     "stack": [115, 116, 117, 108, 109, 110, 111, 112, 120],
     "order": "bottom-to-top (main → leaves)",
     "root": "main",
     "leaves": [120]
   }
   ```

**Impact**:
- Provides authoritative PR stack structure
- Used by `stack-cycle-next-pr` and other tools
- Prevents processing errors from misunderstanding stack order
- Reduces setup time for AI per session

**Effort**: Low (bash/Go, wraps gh commands)

---

#### 3.2 Pre-Commit Hook Remediation (`fix-precommit-failures`)

**Current State**: AI manually fixes pre-commit failures (formatting, linting).

**Workflow Step**: STEP 4 (Remediate CI failures)

**Problem**:
- Pre-commit hook output tells us exactly what to fix
- AI could run the same fixers that the hooks use
- Fixers often have `--fix` / `--write` modes

**Solution**: Create `fix-precommit-failures` script that:
1. Accept list of pre-commit hook failures (from `analyze-ci-failures`)
2. For each hook, apply the fixer:
   - `prettier` → `npx prettier --write <files>`
   - `PSScriptAnalyzer` → invoke Invoke-ScriptAnalyzer with -Fix
   - `trailing-whitespace` → sed to remove trailing spaces
   - `end-of-file-fixer` → add final newline
   - `mixed-line-ending` → dos2unix or git autocrlf
3. Run `pre-commit run --all-files` to verify fixes
4. Return exit code: 0 = all fixed, 1 = some fixes still failing

**Impact**:
- Automates all safe pre-commit fixes
- Eliminates manual formatting work
- Ensures consistent formatting across all PRs
- Reduces AI tokens by ~50-100 per PR with pre-commit issues

**Effort**: Low to Medium (bash, wraps existing tool commands)

---

### PRIORITY 4: Monitoring & Convenience

#### 4.1 Review Readiness Poller (`poll-review-status`)

**Current State**: Manual checking if Copilot review is ready to process.

**Workflow Step**: Between STEP 12 and STEP 13 (after requesting re-review)

**Problem**:
- After requesting review, user must manually wait and check if it's ready
- Currently no automation for "tell me when the review is done"

**Solution**: Create `poll-review-status` script that:
1. Accept PR number and poll interval (default 60 seconds)
2. Loop:
   - Fetch latest review status
   - Check if new comments since last check
   - If found → exit with status "review ready"
   - If timeout (default 30 min) → exit with "timeout"
3. Optional: send desktop notification when ready

**Impact**:
- Allows AI to wait autonomously for reviews
- Enables batch processing ("process PRs #116-120 in sequence, wait for reviews between each")
- Reduces manual intervention points
- Reduces human attention cost

**Effort**: Low (bash, wraps gh commands)

---

#### 4.2 Commit Message Validator (`validate-commit-msg`)

**Current State**: Manual checking that commits follow conventional-commits standard.

**Workflow Step**: STEP 8 (Commit and push)

**Problem**:
- Commit message must follow conventional-commits format
- No automated check before pushing

**Solution**: Create `validate-commit-msg` script that:
1. Accept commit message text
2. Check format: `<type>(<scope>): <subject>`
3. Validate:
   - Type in (feat, fix, docs, style, refactor, perf, test, chore, ci)
   - Scope is relevant (optional)
   - Subject is imperative, < 50 chars
   - Body (if present) is wrapped at 72 chars
4. Return exit code: 0 = valid, 1 = invalid
5. Output detailed feedback for each violation

**Impact**:
- Prevents invalid commit messages before push
- Catches issues early
- Reduces rework from CI or reviewer feedback
- Integrates with commit workflow

**Effort**: Low (bash, simple pattern matching)

---

## Implementation Roadmap

### Phase 1 (Highest ROI - 2-3 days) — ✅ DONE (2026-08-19)
1. ✅ `analyze-pr-stack` (enables all other tools) — handles GitHub's
   post-merge base-retargeting via full graph traversal, not just a linear
   walk; validated against the live owner/repo stack.
2. ✅ `categorize-pr-reviews` (high reuse, enables remediation guide) —
   parses suppressed comments straight out of Copilot review bodies and
   merges them with open review threads; validated against real suppressed
   comments on PRs #108, #109, #116, #117.
3. ✅ `stack-cycle-next-pr` (automates entire cycle-back logic) — composes
   `latest-copilot-review-id`, `is-sentinel-review`, and
   `has-post-sentinel-commits`; accepts PR numbers as args or piped from
   `analyze-pr-stack --json`.

See `README.md` for full usage of each.

### Phase 2 (Core Rebase Automation - 3-4 days) — ✅ DONE (2026-08-19)
4. ✅ `auto-rebase-pr` (highest impact, most complex) — cascades rebase →
   cherry-pick → `apply-pr-diffs`; does NOT auto-push (requires `--push`),
   since force-pushing a shared branch is hard to reverse. Validated with a
   disposable local bare-repo remote through all three strategies, including
   a genuine cherry-pick conflict cascading correctly to the diff-apply
   fallback.
5. ✅ `apply-pr-diffs` (supports auto-rebase) — validated clean-apply,
   pre-commit-failure, and branch-exists-guard paths.

Note: `auto-rebase-pr` tries direct rebase first per this roadmap's original
order. `PR-Review-Resolution-Process.md` instead lists cherry-pick as
"recommended first" and direct rebase as a "legacy approach". In practice a
plain rebase and a cherry-pick of the same commits hit the same conflicts
(same 3-way merge), so trying the cheap rebase first costs nothing when it
works and falls through identically when it doesn't. Revisit the doc's
wording if that assumption turns out wrong in practice.

### Phase 3 (Failure Analysis - 2-3 days)
6. `analyze-ci-failures` (supports fix-precommit-failures)
7. `fix-precommit-failures` (low-hanging fruit)

### Phase 4 (Remediation Support - 2-3 days)
8. `remediation-strategies` (high utility, moderate complexity)

### Phase 5 (Convenience - 1-2 days)
9. `poll-review-status` (nice-to-have, low effort)
10. `validate-commit-msg` (nice-to-have, low effort)

## Estimated Impact

### AI Credits Reduction

| Tool | Per-PR Savings | Frequency | Annual Savings |
|------|-----------------|-----------|-----------------|
| auto-rebase-pr | 500-1000 tokens | ~50 PRs/month | ~300-600K tokens |
| categorize-pr-reviews | 300-500 | ~50 PRs/month | ~180-300K tokens |
| stack-cycle-next-pr | 200-300 | ~10 cycles/month | ~24-36K tokens |
| apply-pr-diffs | 200-400 | ~20 PRs/month | ~48-96K tokens |
| analyze-ci-failures | 200-500 | ~20 PRs/month | ~48-120K tokens |
| fix-precommit-failures | 50-100 | ~40 PRs/month | ~24-48K tokens |
| remediation-strategies | 100-200 | ~50 PRs/month | ~60-120K tokens |
| **Total** | **~1500-3100 tokens/PR** | **~50 PRs/month** | **~684K-1.4M tokens** |

At GitHub Copilot Pro rates (~3M tokens/$20 USD), this represents **$5-10 USD savings per month** or **$60-120 per year** in Copilot credits alone. More importantly, it reduces AI reasoning time by 60-70%, enabling faster PR workflows and better resource allocation.

---

## Success Criteria

- [x] All Phase 1 tools deployed and tested on 5+ real PRs (owner/repo #108, #109, #110, #111, #115, #116, #117, #120)
- [ ] Entire stack workflow (PR #115-#120 cycle) can run with AI only for "remediation logic" steps
- [ ] AI tokens per-PR drop from current ~2000-3000 to target ~500-1000
- [ ] Zero manual "decision making" for rebase, review collection, or cycle-back steps
- [ ] All tools documented with examples in tool README.md

---

## Related Documentation

- `.github/PR-Review-Resolution-Process.md` — workflow to automate
- `~/.copilot/PR-Review-Resolution-Process.md` — user-level copy of workflow
- Session notes: stacked PR review workflow for owner/repo repository
