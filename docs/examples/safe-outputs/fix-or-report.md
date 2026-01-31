# Fix-or-Report Pattern

This example demonstrates a progressive approach where the workflow attempts an automated fix first, and falls back to creating an issue if the fix cannot be automated.

## Pattern Overview

**Use Case:** Dependency updates with automated and manual fallback paths

**Output Strategy:**
- Automated fix possible → Create pull request with changes
- Manual intervention needed → Create issue with details
- Always → Add comment with summary and status

**Benefits:**
- Maximizes automation (fixes what can be fixed)
- Surfaces complex cases for human review
- Complete audit trail with PR or issue
- Reduces manual work for routine updates

## Workflow Implementation

```yaml
---
name: Dependency Update Bot
description: Updates dependencies automatically or creates issues for manual updates
on:
  schedule: weekly
  workflow_dispatch:
permissions:
  contents: write
  pull-requests: write
  issues: write
engine: copilot
tools:
  github:
    toolsets: [default]
  bash:
    - "*"
  edit:
    enabled: true
safe-outputs:
  create-pull-request:
    title-prefix: "[deps] "
    labels: [dependencies, automated]
    draft: false  # Not draft - safe updates
    expires: 14d
    max: 1
  create-issue:
    title-prefix: "[deps-manual] "
    labels: [dependencies, needs-review, manual-update]
    max: 5  # Multiple dependencies may need manual updates
  add-comment:
    target: "*"
    max: 3
timeout-minutes: 30
---

# Dependency Update Bot

You are an automated dependency update bot that attempts to update dependencies safely, creating PRs for successful updates and issues for those requiring manual intervention.

## Objective

Update project dependencies following this strategy:
1. **Analyze** all outdated dependencies
2. **Categorize** by update safety (patch/minor/major, breaking changes)
3. **Attempt** automated update for safe dependencies
4. **Create PR** if update is successful and tests pass
5. **Create Issue** if update requires manual intervention
6. **Comment** with summary of all actions taken

## Update Process

### Step 1: Check for Outdated Dependencies

```bash
# Check npm dependencies
npm outdated --json > /tmp/npm-outdated.json

# Check security vulnerabilities
npm audit --json > /tmp/npm-audit.json

# Parse results
cat /tmp/npm-outdated.json | jq '.'
```

### Step 2: Categorize Updates by Safety

```bash
# Categorize updates
cat /tmp/npm-outdated.json | jq '
  to_entries | 
  map({
    package: .key,
    current: .value.current,
    wanted: .value.wanted,
    latest: .value.latest,
    type: (
      if (.value.current | split(".")[0]) != (.value.latest | split(".")[0])
      then "major"
      elif (.value.current | split(".")[1]) != (.value.latest | split(".")[1])
      then "minor"
      else "patch"
      end
    )
  }) |
  group_by(.type)
' > /tmp/categorized-updates.json
```

### Step 3: Attempt Automated Update

Try to update safe dependencies (patch and minor versions):

```bash
# Update patch versions (safest)
cat /tmp/categorized-updates.json | jq -r '
  .[] | 
  select(.[0].type == "patch") | 
  .[].package
' | xargs -I {} npm install {}@latest

# Run tests to verify updates
npm test

# Check if tests pass
if [ $? -eq 0 ]; then
  echo "✅ Tests passed - safe to create PR"
  UPDATE_SUCCESSFUL=true
else
  echo "❌ Tests failed - requires manual review"
  UPDATE_SUCCESSFUL=false
fi
```

### Step 4: Decision Point - PR or Issue?

#### If Update Successful → Create Pull Request

```markdown
[create-pull-request]
title: Update dependencies ({package_count} packages)
body: |
  ## Dependency Updates
  
  This PR updates {package_count} dependencies to their latest compatible versions.
  
  ### Updated Packages
  
  | Package | From | To | Type |
  |---------|------|-----|------|
  {package_update_table}
  
  ### Safety Analysis
  
  ✅ **All updates are patch or minor versions**
  - No breaking changes expected
  - All tests passing
  - No security vulnerabilities introduced
  
  ### Test Results
  
  ```
  {test_results_summary}
  ```
  
  All {test_count} tests passed successfully.
  
  ### Security Impact
  
  {if security_fixes_count > 0}
  🔒 This update includes {security_fixes_count} security fixes:
  {security_fixes_list}
  {else}
  No security vulnerabilities in updated packages.
  {endif}
  
  ### Changes Made
  
  - Updated `package.json`
  - Updated `package-lock.json`
  - Verified tests pass
  - Checked for breaking changes
  
  ### Review Checklist
  
  - [x] Tests pass
  - [x] No breaking changes
  - [x] Dependencies compatible
  - [ ] Visual/manual testing (if applicable)
  
  ### Rollback Plan
  
  If issues arise, rollback with:
  ```bash
  git revert {commit_sha}
  npm install
  ```
  
  ## Metadata
  
  - **Update Date:** {update_date}
  - **Workflow Run:** {workflow_run_url}
  - **Previous Versions:** See commit diff
[/create-pull-request]
```

#### If Update Failed → Create Issue

```markdown
[create-issue]
title: Manual dependency update required: {package_name}
body: |
  ## Overview
  
  The dependency `{package_name}` has a new version available but requires manual intervention to update.
  
  ### Update Details
  
  - **Package:** {package_name}
  - **Current Version:** {current_version}
  - **Latest Version:** {latest_version}
  - **Update Type:** {major|minor|patch}
  
  ### Why Manual Update Required
  
  {reason_for_manual_update}
  
  **Common reasons:**
  - ⚠️ Major version change (breaking changes likely)
  - ❌ Tests fail after update
  - 🔄 Requires code changes to accommodate new API
  - 📚 Migration guide needed
  - 🔗 Peer dependency conflicts
  
  ### Test Failure Details
  
  {if tests_failed}
  ```
  {test_failure_output}
  ```
  {endif}
  
  ### Breaking Changes
  
  {if breaking_changes_detected}
  Review the changelog for breaking changes:
  {changelog_url}
  
  **Known breaking changes:**
  {breaking_changes_list}
  {endif}
  
  ### Migration Steps
  
  1. **Review changelog:**
     {changelog_url}
  
  2. **Update package.json:**
     ```json
     "{package_name}": "^{latest_version}"
     ```
  
  3. **Install and test:**
     ```bash
     npm install
     npm test
     ```
  
  4. **Address breaking changes:**
     - Update deprecated API usage
     - Modify code to match new patterns
     - Update tests if needed
  
  5. **Verify in development:**
     - Run application locally
     - Test critical paths
     - Check for console errors/warnings
  
  ### Impact Assessment
  
  **Priority:** {high|medium|low}
  
  {if has_security_vulnerability}
  🔴 **SECURITY FIX** - This update includes security fixes for:
  {security_vulnerability_details}
  
  **Recommended:** Update within {timeframe}
  {endif}
  
  {if no_security_issues}
  This is a routine update with no immediate security concerns.
  {endif}
  
  ### Additional Context
  
  - **Dependencies affected:** {affected_dependencies_count}
  - **Required code changes:** {estimated_change_scope}
  - **Estimated effort:** {estimated_hours} hours
  
  ### Related Items
  
  - Dependency audit: #{audit_discussion_number}
  - Previous updates: {related_prs}
  - Upstream issue: {upstream_issue_url}
  
  ### Acceptance Criteria
  
  - [ ] Package updated to {latest_version}
  - [ ] All tests passing
  - [ ] No breaking changes in application
  - [ ] Code changes reviewed
  - [ ] PR created and merged
  
  ## Metadata
  
  - **Detection Date:** {detection_date}
  - **Workflow Run:** {workflow_run_url}
  - **Automated Update Attempted:** Yes (failed)
[/create-issue]
```

### Step 5: Add Summary Comment

Always create a summary comment, regardless of PR or issue creation:

```markdown
[add-comment]
target: {related_issue_or_pr}  # Or create new discussion
body: |
  ## 📦 Dependency Update Summary
  
  Analyzed {total_packages} dependencies for updates.
  
  ### Actions Taken
  
  {if pr_created}
  ✅ **Pull Request Created:** #{pr_number}
  - Updated {updated_count} packages successfully
  - All tests passing
  - Safe to review and merge
  {endif}
  
  {if issues_created}
  ⚠️ **Manual Updates Required:** {issue_count} packages
  {list_of_created_issues}
  - Require code changes or review
  - See individual issues for details
  {endif}
  
  {if no_updates}
  ✅ All dependencies are up to date
  {endif}
  
  ### Summary Table
  
  | Status | Count | Action |
  |--------|-------|--------|
  | Automated | {automated_count} | PR #{pr_number} |
  | Manual | {manual_count} | Issues created |
  | Up to date | {uptodate_count} | No action |
  | Skipped | {skipped_count} | See notes |
  
  ### Security Status
  
  {if security_vulnerabilities_fixed}
  🔒 Security fixes included in PR #{pr_number}
  {endif}
  
  {if security_vulnerabilities_remain}
  ⚠️ Security vulnerabilities require manual updates: {security_issues_list}
  {endif}
  
  ### Next Steps
  
  {if pr_created}
  1. Review and approve PR #{pr_number}
  2. Merge after approval
  3. Monitor production
  {endif}
  
  {if issues_created}
  1. Review manual update issues
  2. Prioritize by security impact
  3. Address breaking changes
  {endif}
  
  ## Full Report
  
  See complete dependency audit: #{audit_discussion_number}
[/add-comment]
```

## Decision Logic

### When to Create PR vs Issue

```javascript
// Pseudocode for decision logic

for (const dependency of outdatedDependencies) {
  const updateType = classifyUpdate(dependency);
  
  if (isSafeUpdate(dependency)) {
    // Attempt automated update
    const updated = await updateDependency(dependency);
    const testsPass = await runTests();
    
    if (updated && testsPass) {
      // Safe to automate - create PR
      await createPullRequest({
        title: `Update ${dependency.name}`,
        changes: dependency.changes
      });
    } else {
      // Tests failed - needs manual review
      await createIssue({
        title: `Manual update required: ${dependency.name}`,
        reason: 'Tests failed after update'
      });
    }
  } else {
    // Not safe to automate - create issue
    await createIssue({
      title: `Manual update required: ${dependency.name}`,
      reason: determineReason(dependency)
    });
  }
}

function isSafeUpdate(dependency) {
  return (
    dependency.updateType === 'patch' ||
    (dependency.updateType === 'minor' && !hasBreakingChanges(dependency)) &&
    !hasSecurityVulnerabilities(dependency, 'high')
  );
}
```

### Safety Criteria

| Criteria | Safe for PR | Needs Issue |
|----------|-------------|-------------|
| Patch update (0.0.X) | ✅ Yes | - |
| Minor update (0.X.0) | ✅ Yes (if no breaking changes) | ⚠️ If breaking changes |
| Major update (X.0.0) | ❌ No | ✅ Yes |
| Tests pass after update | ✅ Yes | - |
| Tests fail after update | ❌ No | ✅ Yes |
| Peer dependency conflicts | ❌ No | ✅ Yes |
| Deprecated API usage | ❌ No | ✅ Yes |
| Security vulnerability | ⚠️ If patch available | ✅ If major update needed |

## Success Metrics

- **Automation Rate:** % of updates handled via PR
- **Time Savings:** Manual updates identified quickly
- **Test Coverage:** All updates validated by tests
- **Audit Trail:** Complete history via PR or issue

## Variations

### Pattern 3A: Security-First

Always prioritize security fixes:

```markdown
If security vulnerability:
  - Critical (CVSS ≥ 9.0) → Create PR immediately, notify team
  - High (CVSS 7.0-8.9) → Create PR or issue
  - Medium/Low → Include in regular updates
```

### Pattern 3B: Staged Rollout

Create PR for staging environment first:

```markdown
1. Create PR targeting `staging` branch
2. Deploy to staging
3. Monitor for issues
4. If successful, create PR for `main`
5. If issues, create issue for investigation
```

### Pattern 3C: Dependency Groups

Group related dependencies:

```markdown
- Testing dependencies → One PR
- Build tools → One PR  
- Runtime dependencies → Individual PRs (more risk)
```

## Related Patterns

- [Conditional Output](./conditional-output.md) - For routing by severity
- [Multi-Output Analysis](./multi-output-analysis.md) - For comprehensive reports
- [Comment Pattern](./comment-pattern.md) - For status updates

---

**Pattern Type:** Fix-or-Report (Progressive)  
**Complexity:** High  
**Use Cases:** Dependency updates, automated refactoring, configuration updates  
**Related Workflows:** `breaking-change-checker.md`, `ci-coach.md`
