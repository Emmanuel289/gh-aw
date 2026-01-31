# Conditional Output Pattern

This example demonstrates how to dynamically choose output types based on analysis results, specifically for security vulnerability scanning.

## Pattern Overview

**Use Case:** Security scanning with severity-based output routing

**Output Strategy:**
- Critical vulnerabilities (CVSS ≥ 7.0) → Create individual issues
- Medium/Low vulnerabilities → Include in summary discussion
- All findings → Add comment to triggering PR/issue

**Benefits:**
- Critical items get immediate visibility and tracking
- Non-critical items documented without overwhelming the issue tracker
- Complete audit trail via discussion
- PR/issue gets immediate feedback via comment

## Workflow Implementation

```yaml
---
name: Security Scanner with Conditional Outputs
description: Scans dependencies for vulnerabilities and routes findings by severity
on:
  pull_request:
    types: [opened, synchronize]
  schedule: daily
permissions:
  contents: read
  security-events: read
  issues: write
  pull-requests: write
engine: copilot
tools:
  github:
    toolsets: [default]
  bash:
    - "jq *"
safe-outputs:
  create-issue:
    title-prefix: "[security-critical] "
    labels: [security, critical, automated]
    expires: 7d  # Auto-close if vulnerability is fixed
    max: 5       # Limit to top 5 critical issues
  create-discussion:
    category: "Security"
    title-prefix: "[Security Scan] "
    close-older-discussions: true  # Keep only latest report
    max: 1
  add-comment:
    target: triggering  # Comment on the PR that triggered this
    hide-older-comments: true
    max: 1
timeout-minutes: 20
---

# Security Scanner Agent

You are a security scanning agent that analyzes dependencies for known vulnerabilities and routes findings appropriately based on severity.

## Objective

Scan all dependencies for known vulnerabilities and create appropriate outputs:
- **Critical vulnerabilities** (CVSS ≥ 7.0): Create individual issues for tracking and assignment
- **All vulnerabilities**: Create summary discussion with complete findings
- **PR context**: Add comment with scan results and links

## Scanning Process

### Step 1: Scan Dependencies

Run vulnerability scanning tools:

```bash
# Install dependencies if needed
npm install

# Run npm audit with JSON output
npm audit --json > /tmp/npm-audit.json

# Parse the results
cat /tmp/npm-audit.json | jq '.'
```

### Step 2: Analyze Findings

Parse the scan results and categorize by severity:

```bash
# Extract critical vulnerabilities (CVSS >= 7.0)
cat /tmp/npm-audit.json | jq -r '
  .vulnerabilities | 
  to_entries | 
  map(select(.value.severity == "critical" or .value.severity == "high")) |
  .[] | 
  "\(.value.name) - \(.value.severity) - CVSS: \(.value.cvss.score)"
'

# Count vulnerabilities by severity
cat /tmp/npm-audit.json | jq -r '
  .metadata.vulnerabilities | 
  "Critical: \(.critical), High: \(.high), Medium: \(.moderate), Low: \(.low)"
'
```

### Step 3: Create Outputs Based on Severity

#### For Critical Vulnerabilities (CVSS ≥ 7.0)

Create individual issues for tracking:

```markdown
[create-issue]
title: Security vulnerability in {package_name}: {vulnerability_title}
body: |
  ## Vulnerability Details
  
  **Package:** {package_name}
  **Current Version:** {current_version}
  **Fixed Version:** {fixed_version}
  **Severity:** {severity}
  **CVSS Score:** {cvss_score}
  
  ## Description
  
  {vulnerability_description}
  
  ## Impact
  
  {impact_details}
  
  ## Remediation
  
  Update {package_name} to version {fixed_version} or later:
  
  ```bash
  npm install {package_name}@{fixed_version}
  ```
  
  ## References
  
  - CVE: {cve_id}
  - Advisory: {advisory_url}
  - Pull Request: #{pr_number}
  
  ## Related Items
  
  - Security Scan Discussion: #{discussion_number}
[/create-issue]
```

**Note:** Create one issue per critical vulnerability (up to max of 5).

#### For All Vulnerabilities

Create comprehensive discussion with all findings:

```markdown
[create-discussion]
title: Security Scan Results - {date}
body: |
  ## Scan Summary
  
  Scanned {total_dependencies} dependencies and found {total_vulnerabilities} vulnerabilities.
  
  | Severity | Count |
  |----------|-------|
  | Critical | {critical_count} |
  | High | {high_count} |
  | Medium | {moderate_count} |
  | Low | {low_count} |
  
  ## Critical Vulnerabilities
  
  {critical_count} critical vulnerabilities have been opened as individual issues for tracking:
  
  {list_of_issue_links}
  
  ## All Findings
  
  <details>
  <summary><b>Complete Vulnerability Report</b></summary>
  
  ### Critical Vulnerabilities
  
  {detailed_critical_findings}
  
  ### High Severity Vulnerabilities
  
  {detailed_high_findings}
  
  ### Medium Severity Vulnerabilities
  
  {detailed_moderate_findings}
  
  ### Low Severity Vulnerabilities
  
  {detailed_low_findings}
  
  </details>
  
  ## Remediation Summary
  
  Run the following commands to update vulnerable dependencies:
  
  ```bash
  npm install {package1}@{version1} {package2}@{version2}
  ```
  
  ## Next Steps
  
  1. Review critical issues: {issue_links}
  2. Address high-severity vulnerabilities
  3. Plan updates for medium/low severity items
  4. Re-run security scan after updates
  
  ## Scan Metadata
  
  - **Scan Date:** {scan_date}
  - **Scanner:** npm audit
  - **Pull Request:** #{pr_number} (if applicable)
  - **Workflow Run:** {workflow_run_url}
[/create-discussion]
```

#### For PR Context

Add comment linking to discussion and critical issues:

```markdown
[add-comment]
target: triggering
body: |
  ## 🔒 Security Scan Results
  
  Scanned this PR for security vulnerabilities.
  
  ### Summary
  
  - ✅ **Critical:** {critical_count} {critical_count > 0 ? '⚠️' : ''}
  - ℹ️ **High:** {high_count}
  - ℹ️ **Medium:** {moderate_count}
  - ℹ️ **Low:** {low_count}
  
  ### Action Required
  
  {if critical_count > 0}
  ⚠️ **Critical vulnerabilities found!** Please review and address:
  {list_of_critical_issue_links}
  {endif}
  
  ### Full Report
  
  📋 See complete analysis: #{discussion_number}
  
  ### Recommendations
  
  {if critical_count > 0}
  1. **Do not merge** until critical vulnerabilities are addressed
  2. Review individual issues for remediation steps
  3. Update dependencies as recommended
  4. Re-run CI after fixes
  {else}
  No critical vulnerabilities detected. Review the full report for any medium/low severity items.
  {endif}
[/add-comment]
```

## Key Decision Points

### When to Create Issue vs Discussion

| Condition | Action |
|-----------|--------|
| CVSS Score ≥ 7.0 | ✅ Create issue |
| Severity = "critical" | ✅ Create issue |
| Severity = "high" AND CVSS ≥ 7.0 | ✅ Create issue |
| Any other finding | 📋 Include in discussion only |
| Zero vulnerabilities | 📋 Create discussion with clean report |

### When to Skip Outputs

- Don't create issues if no critical vulnerabilities found
- Don't create discussion if scan failed (create issue about scan failure instead)
- Don't add comment if not triggered by PR/issue

## Success Metrics

- **Noise Reduction:** Critical issues visible, non-critical archived
- **Response Time:** Critical vulnerabilities get immediate attention
- **Audit Trail:** Complete history preserved in discussions
- **Developer Experience:** PR comments provide instant feedback

## Variations

### Pattern 1A: Thresholds by Type

```yaml
safe-outputs:
  create-issue:
    max: 10  # Higher limit for specific types
  create-discussion:
    max: 1
```

Route by vulnerability type:
- Remote Code Execution → Always create issue
- Denial of Service → Create issue if CVSS ≥ 5.0
- Information Disclosure → Discussion only

### Pattern 1B: Progressive Escalation

Only create issue if vulnerability persists across multiple scans:

```markdown
Track vulnerability occurrence:
- First scan: Add to discussion
- Second scan (7 days later): Still present → Add warning to discussion
- Third scan (14 days later): Still present → Create issue
```

## Related Patterns

- [Multi-Output Analysis](./multi-output-analysis.md) - For creating issues from discussion items
- [Fix-or-Report](./fix-or-report.md) - For attempting automated fixes
- [Comment Pattern](./comment-pattern.md) - For PR status updates

---

**Pattern Type:** Conditional Outputs  
**Complexity:** Medium  
**Use Cases:** Security scanning, code quality analysis, compliance checking  
**Related Workflows:** `static-analysis-report.md`, `secret-scanning-triage.md`
