# Comment Pattern

This example demonstrates a comment-first approach for status updates with escalation to issues only when necessary.

## Pattern Overview

**Use Case:** CI status reporting with escalation for persistent failures

**Output Strategy:**
- Always → Add comment with CI results to PR/issue
- Persistent failure (3+ runs) → Create issue for investigation
- Hide older comments to reduce noise

**Benefits:**
- All status updates in one place
- Historical context preserved (collapsed older comments)
- Escalates only genuine problems
- Reduces issue tracker noise

## Workflow Implementation

```yaml
---
name: CI Status Reporter
description: Reports CI results and escalates persistent failures
on:
  pull_request:
    types: [opened, synchronize]
  push:
    branches: [main, develop]
permissions:
  contents: read
  pull-requests: write
  issues: write
  checks: read
engine: copilot
tools:
  github:
    toolsets: [default, actions]
  bash:
    - "jq *"
safe-outputs:
  add-comment:
    target: triggering
    hide-older-comments: true  # Hide previous CI comments
    max: 1
  create-issue:
    title-prefix: "[ci-failure] "
    labels: [ci, bug, needs-investigation]
    max: 1  # Only create issue for persistent failures
timeout-minutes: 15
---

# CI Status Reporter

You are a CI status reporter that provides detailed feedback on continuous integration runs and escalates persistent failures.

## Objective

Monitor CI status and provide feedback:
1. **Always** → Add comment with CI results to the PR
2. **If persistent failure** (3+ consecutive failures) → Create issue
3. **Hide older comments** → Keep comment thread clean

## Monitoring Process

### Step 1: Fetch CI Results

```bash
# Get workflow runs for this commit
gh api "/repos/${{ github.repository }}/actions/runs" \
  --jq ".workflow_runs[] | select(.head_sha == \"${{ github.sha }}\")" \
  > /tmp/workflow-runs.json

# Get check runs for detailed status
gh api "/repos/${{ github.repository }}/commits/${{ github.sha }}/check-runs" \
  > /tmp/check-runs.json

# Parse results
cat /tmp/workflow-runs.json | jq '.[] | {name, status, conclusion, html_url}'
```

### Step 2: Analyze Failure Patterns

```bash
# Check for persistent failures (last 3 runs)
gh api "/repos/${{ github.repository }}/actions/workflows/{workflow_id}/runs" \
  --jq ".workflow_runs[:3] | map(.conclusion) | all(. == \"failure\")" \
  > /tmp/persistent-failure.txt

PERSISTENT_FAILURE=$(cat /tmp/persistent-failure.txt)

# Get failure count from recent history
FAILURE_COUNT=$(gh api "/repos/${{ github.repository }}/actions/workflows/{workflow_id}/runs" \
  --jq '.workflow_runs[:10] | map(select(.conclusion == "failure")) | length')
```

### Step 3: Add Status Comment

Always add a comment with the current CI status:

```markdown
[add-comment]
target: triggering
body: |
  ## 🔄 CI Status Report
  
  **Status:** {status_emoji} {status_text}
  
  ### Summary
  
  | Workflow | Status | Duration | Details |
  |----------|--------|----------|---------|
  {workflow_status_table}
  
  ### {if all_passed}✅ All Checks Passed{/if}{if some_failed}❌ Some Checks Failed{/if}{if all_failed}🚫 All Checks Failed{/if}
  
  {if all_passed}
  All CI checks completed successfully! This PR is ready for review.
  
  **Next Steps:**
  - [ ] Code review
  - [ ] Approval from maintainers
  - [ ] Merge when ready
  {/if}
  
  {if some_failed}
  Some CI checks failed. Please review the failures below.
  
  #### Failed Checks
  
  {failed_checks_details}
  
  **Common Fixes:**
  - Check test failures in the logs
  - Verify linting and formatting
  - Ensure all dependencies are installed
  - Review security scan results
  {/if}
  
  {if all_failed}
  ⚠️ **All CI checks failed!**
  
  This may indicate a significant issue. Please investigate immediately.
  
  {if persistent_failure}
  🔴 **Persistent Failure Detected**
  
  This workflow has failed {failure_count} consecutive times. An issue has been created for investigation: #{issue_number}
  {/if}
  {/if}
  
  ### Detailed Results
  
  <details>
  <summary><b>Test Results</b></summary>
  
  {test_results_summary}
  
  **Failed Tests:**
  {failed_tests_list}
  
  </details>
  
  <details>
  <summary><b>Linting Results</b></summary>
  
  {linting_results}
  
  </details>
  
  <details>
  <summary><b>Build Logs</b></summary>
  
  {build_log_excerpt}
  
  Full logs: {workflow_run_url}
  
  </details>
  
  ### Performance Metrics
  
  - **Total Duration:** {total_duration}
  - **Fastest Job:** {fastest_job} ({fastest_duration})
  - **Slowest Job:** {slowest_job} ({slowest_duration})
  
  {if duration_increased}
  ⚠️ **Performance Note:** CI duration increased by {duration_increase} compared to previous run.
  {/if}
  
  ### Quick Actions
  
  {if can_rerun}
  - 🔄 [Re-run failed jobs]({rerun_url})
  {/if}
  - 📊 [View detailed logs]({logs_url})
  - 📈 [View workflow history]({history_url})
  
  ### Comparison with Previous Run
  
  {if has_previous_run}
  | Metric | Current | Previous | Change |
  |--------|---------|----------|--------|
  | Status | {current_status} | {previous_status} | {status_change} |
  | Duration | {current_duration} | {previous_duration} | {duration_change} |
  | Failed Tests | {current_failed} | {previous_failed} | {test_change} |
  {/if}
  
  ---
  
  *Last updated: {timestamp}*  
  *Workflow: [{workflow_name}]({workflow_run_url})*
[/add-comment]
```

**Note:** Use `hide-older-comments: true` to automatically hide previous CI comments.

### Step 4: Escalate Persistent Failures

If failure persists for 3+ consecutive runs, create an issue:

```markdown
[create-issue]
title: Persistent CI failure in {workflow_name} workflow
body: |
  ## 🚨 Persistent CI Failure Alert
  
  The `{workflow_name}` workflow has failed **{failure_count} consecutive times**.
  
  ### Failure Summary
  
  - **Workflow:** {workflow_name}
  - **First Failure:** {first_failure_date}
  - **Latest Failure:** {latest_failure_date}
  - **Affected Branch:** {branch_name}
  - **Failure Count:** {failure_count} consecutive failures
  
  ### Recent Failure History
  
  | Run | Date | Duration | Conclusion |
  |-----|------|----------|------------|
  {failure_history_table}
  
  ### Common Failure Pattern
  
  {if consistent_failure_reason}
  All failures show the same root cause:
  
  ```
  {error_message}
  ```
  
  **Likely Cause:** {likely_cause}
  {else}
  Failures show different error messages. This may indicate:
  - Flaky tests
  - Environment issues
  - Race conditions
  - External dependency problems
  {/if}
  
  ### Failed Jobs
  
  {failed_jobs_details}
  
  ### Failing Tests
  
  {if has_test_failures}
  The following tests are consistently failing:
  
  {failing_tests_list}
  {/if}
  
  ### Error Analysis
  
  <details>
  <summary><b>Error Logs</b></summary>
  
  {error_logs}
  
  </details>
  
  <details>
  <summary><b>Stack Traces</b></summary>
  
  {stack_traces}
  
  </details>
  
  ### Impact Assessment
  
  **Severity:** {high|medium|low}
  
  {if blocks_deployment}
  🔴 **BLOCKING** - This failure is preventing deployments
  {/if}
  
  {if blocks_merges}
  ⚠️ **IMPACTING** - This failure may be blocking PR merges
  {/if}
  
  **Affected Areas:**
  - {affected_module_1}
  - {affected_module_2}
  
  ### Recommended Actions
  
  1. **Immediate Investigation**
     - Review error logs: {logs_url}
     - Check for environment changes
     - Verify external dependencies
  
  2. **Testing**
     - Reproduce failure locally
     - Run failing tests in isolation
     - Check for race conditions
  
  3. **Resolution**
     - Fix identified issues
     - Add regression tests
     - Update CI configuration if needed
  
  4. **Verification**
     - Re-run CI pipeline
     - Monitor next 3 runs
     - Close issue when resolved
  
  ### Debug Information
  
  - **Repository:** ${{ github.repository }}
  - **Branch:** {branch_name}
  - **Commit SHA:** {commit_sha}
  - **Workflow File:** `.github/workflows/{workflow_file}`
  - **Runner OS:** {runner_os}
  - **Node Version:** {node_version}
  
  ### Related PRs
  
  {if has_related_prs}
  This failure is affecting the following PRs:
  {list_of_affected_prs}
  {/if}
  
  ### Historical Context
  
  {if has_similar_past_failures}
  Similar failures occurred previously:
  {list_of_similar_issues}
  {/if}
  
  ### Acceptance Criteria
  
  - [ ] Root cause identified
  - [ ] Fix implemented
  - [ ] CI passes 3 consecutive times
  - [ ] Regression tests added
  - [ ] Documentation updated (if needed)
  
  ## Metadata
  
  - **Detection Date:** {detection_date}
  - **Latest Workflow Run:** {workflow_run_url}
  - **Failure Threshold:** 3 consecutive failures
  
  ---
  
  *Automatically created by CI Status Reporter*  
  *Related PR: #{pr_number}*
[/create-issue]
```

## Decision Logic

### When to Comment vs Create Issue

```javascript
// Pseudocode for escalation logic

const recentRuns = await getRecentWorkflowRuns(10);
const consecutiveFailures = countConsecutiveFailures(recentRuns);

// Always add comment with status
await addComment({
  target: 'triggering',
  body: formatCIStatus(currentRun)
});

// Escalate if persistent failure
if (consecutiveFailures >= 3 && !hasExistingIssue()) {
  await createIssue({
    title: `Persistent CI failure in ${workflowName}`,
    body: formatFailureReport(recentRuns)
  });
}

// Close issue if now passing
if (currentRun.conclusion === 'success' && hasExistingIssue()) {
  await addComment({
    target: existingIssueNumber,
    body: '✅ CI is now passing! Closing this issue.'
  });
  await closeIssue(existingIssueNumber);
}
```

### Escalation Criteria

| Scenario | Add Comment | Create Issue | Close Issue |
|----------|-------------|--------------|-------------|
| First failure | ✅ Yes | ❌ No | - |
| Second consecutive failure | ✅ Yes | ❌ No | - |
| Third consecutive failure | ✅ Yes | ✅ Yes | - |
| Success after failures | ✅ Yes | ❌ No | ✅ Yes (if exists) |
| Flaky test (intermittent) | ✅ Yes | ⚠️ Maybe (track pattern) | - |

### Comment Visibility Strategy

**Hide Older Comments:**
```yaml
safe-outputs:
  add-comment:
    hide-older-comments: true  # Auto-hides previous comments
```

This keeps only the latest status visible, with older comments collapsed to reduce noise.

**Allowed Hide Reasons:**
- `outdated` - Status is superseded by newer run
- `resolved` - Issue was fixed
- `duplicate` - Multiple comments for same run

## Success Metrics

- **Clean Thread:** Only latest status visible
- **Quick Escalation:** Persistent issues identified immediately
- **Low Noise:** Issues created only for genuine problems
- **Fast Resolution:** Clear action items in both comments and issues

## Variations

### Pattern 4A: Multi-Stage Escalation

```markdown
- 1st failure → Comment with status
- 2nd failure → Comment with warning
- 3rd failure → Create issue
- 5th failure → Notify team (mention @team)
```

### Pattern 4B: Performance Degradation Tracking

Track CI performance over time:

```markdown
[add-comment]
body: |
  ⚠️ **Performance Alert**
  
  CI duration has increased 30% over last 7 days:
  - 7 days ago: 5m 30s
  - Today: 7m 15s
  
  Consider investigating slow jobs.
```

### Pattern 4C: Security Scan Results

For security scans, always comment but only create issue for critical findings:

```markdown
[add-comment]
target: triggering
body: Security scan complete. {vulnerability_count} vulnerabilities found.

[create-issue]  # Only if critical vulnerabilities
title: Critical security vulnerabilities detected
condition: has_critical_vulnerabilities
```

## Related Patterns

- [Conditional Output](./conditional-output.md) - For routing by severity
- [Multi-Output Analysis](./multi-output-analysis.md) - For comprehensive reports
- [Fix-or-Report](./fix-or-report.md) - For automated fixes

---

**Pattern Type:** Comment-First (Update-Focused)  
**Complexity:** Low-Medium  
**Use Cases:** CI status, PR reviews, progress updates, monitoring  
**Related Workflows:** `ci-coach.md`, `breaking-change-checker.md`, `auto-triage-issues.md`
