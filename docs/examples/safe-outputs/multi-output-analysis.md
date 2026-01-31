# Multi-Output Analysis Pattern

This example demonstrates a hierarchical output strategy where a parent discussion contains the overall analysis, with child issues for actionable sub-items.

## Pattern Overview

**Use Case:** Code quality analysis with actionable findings

**Output Strategy:**
- Discussion → Comprehensive analysis and trends
- Issues → Individual actionable items from analysis
- Comment → Summary with links to discussion and issues

**Benefits:**
- Complete view in one place (discussion)
- Individual tracking for each finding (issues)
- Easy assignment and closure of sub-tasks
- Historical trend tracking via discussions

## Workflow Implementation

```yaml
---
name: Code Quality Analyzer
description: Analyzes code quality metrics and creates actionable issues for improvements
on:
  schedule: weekly
  workflow_dispatch:
permissions:
  contents: read
  issues: write
engine: copilot
tools:
  github:
    toolsets: [default]
  bash:
    - "jq *"
    - "grep *"
safe-outputs:
  create-discussion:
    category: "Reports"
    title-prefix: "[Code Quality] "
    close-older-discussions: true
    max: 1
  create-issue:
    title-prefix: "[quality] "
    labels: [code-quality, automated, needs-triage]
    max: 10  # Limit to top 10 issues
  add-comment:
    target: "*"  # Can comment on any issue/PR
    max: 5
timeout-minutes: 30
---

# Code Quality Analyzer

You are a code quality analyzer that identifies technical debt, code smells, and improvement opportunities.

## Objective

Analyze the codebase for quality issues and create:
1. **Discussion** with complete analysis, trends, and recommendations
2. **Issues** for each actionable finding (limit to top 10 priorities)
3. **Comments** on related PRs/issues for context

## Analysis Process

### Step 1: Run Quality Analysis Tools

```bash
# Run ESLint (JavaScript/TypeScript)
npx eslint . --format json --output-file /tmp/eslint-results.json || true

# Run SonarQube Scanner (if configured)
# sonar-scanner -Dsonar.projectKey=project -Dsonar.host.url=http://localhost:9000

# Analyze complexity
npx complexity-report --format json src/ > /tmp/complexity.json || true

# Check test coverage
npm run test:coverage -- --json > /tmp/coverage.json || true

# Analyze dependencies
npm audit --json > /tmp/audit.json || true
npx depcheck --json > /tmp/depcheck.json || true
```

### Step 2: Categorize Findings

Parse and categorize the results:

```bash
# High priority issues (complexity > 20, critical lint errors, etc.)
cat /tmp/complexity.json | jq '[.[] | select(.complexity > 20)] | length'

# Medium priority (moderate complexity, warnings, etc.)
cat /tmp/eslint-results.json | jq '[.[] | .messages[] | select(.severity == 1)] | length'

# Technical debt indicators
cat /tmp/eslint-results.json | jq '[.[] | .messages[] | select(.message | contains("TODO") or contains("FIXME"))] | length'
```

### Step 3: Create Discussion with Complete Analysis

```markdown
[create-discussion]
title: Code Quality Report - Week {week_number} {year}
body: |
  ## Executive Summary
  
  This week's code quality analysis reveals {total_issues} findings across {categories_count} categories.
  
  ### Key Metrics
  
  | Metric | Current | Previous | Trend |
  |--------|---------|----------|-------|
  | Complexity Score | {current_complexity} | {previous_complexity} | {trend_emoji} |
  | Test Coverage | {current_coverage}% | {previous_coverage}% | {trend_emoji} |
  | Lint Errors | {current_errors} | {previous_errors} | {trend_emoji} |
  | Technical Debt | {debt_hours}h | {previous_debt}h | {trend_emoji} |
  
  ### Priority Distribution
  
  - 🔴 **High Priority:** {high_count} findings
  - 🟡 **Medium Priority:** {medium_count} findings
  - 🟢 **Low Priority:** {low_count} findings
  
  ## Actionable Issues Created
  
  The following issues have been created for high-priority findings:
  
  {list_of_created_issues_with_links}
  
  ## Detailed Analysis
  
  <details>
  <summary><b>High Priority Findings</b></summary>
  
  ### Complexity Hotspots
  
  The following files have cyclomatic complexity > 20:
  
  | File | Function | Complexity | Recommendation |
  |------|----------|------------|----------------|
  {complexity_hotspots_table}
  
  ### Critical Lint Errors
  
  {critical_lint_errors_list}
  
  ### Test Coverage Gaps
  
  {coverage_gaps_list}
  
  </details>
  
  <details>
  <summary><b>Medium Priority Findings</b></summary>
  
  {medium_priority_details}
  
  </details>
  
  <details>
  <summary><b>Technical Debt Analysis</b></summary>
  
  ### TODO/FIXME Comments
  
  Found {todo_count} TODO and {fixme_count} FIXME comments:
  
  {todo_fixme_list}
  
  ### Deprecated API Usage
  
  {deprecated_api_usage}
  
  ### Unused Dependencies
  
  {unused_dependencies_list}
  
  </details>
  
  ## Trends Over Time
  
  ### Complexity Trend (Last 4 Weeks)
  
  ```
  Week {w1}: {complexity_w1}
  Week {w2}: {complexity_w2}
  Week {w3}: {complexity_w3}
  Week {w4}: {complexity_w4}
  ```
  
  ### Coverage Trend (Last 4 Weeks)
  
  ```
  Week {w1}: {coverage_w1}%
  Week {w2}: {coverage_w2}%
  Week {w3}: {coverage_w3}%
  Week {w4}: {coverage_w4}%
  ```
  
  ## Recommendations
  
  1. **Immediate Actions**
     - Address high-priority issues: {high_priority_issues_links}
     - Focus on complexity hotspots in critical paths
  
  2. **Short-term Improvements** (1-2 weeks)
     - Increase test coverage in {low_coverage_modules}
     - Resolve critical lint errors
     - Clean up TODO/FIXME comments
  
  3. **Long-term Strategy**
     - Establish complexity thresholds in CI
     - Implement automated code quality gates
     - Regular refactoring sprints
  
  ## Comparison with Previous Report
  
  Previous report: #{previous_discussion_number}
  
  **Changes since last report:**
  - Complexity: {complexity_change}
  - Coverage: {coverage_change}
  - Errors: {errors_change}
  
  ## Metadata
  
  - **Analysis Date:** {analysis_date}
  - **Branch:** {branch}
  - **Commit:** {commit_sha}
  - **Tools:** ESLint, complexity-report, Jest
  - **Workflow Run:** {workflow_run_url}
[/create-discussion]
```

### Step 4: Create Individual Issues for High-Priority Items

Create one issue per high-priority finding:

```markdown
[create-issue]
title: Reduce complexity in {file_name}::{function_name} (complexity: {score})
body: |
  ## Problem
  
  The function `{function_name}` in `{file_path}` has a cyclomatic complexity of **{complexity_score}**, which exceeds the recommended threshold of 20.
  
  ## Current State
  
  ```{language}
  {code_snippet}
  ```
  
  **Complexity Analysis:**
  - Cyclomatic Complexity: {complexity_score}
  - Number of Branches: {branch_count}
  - Lines of Code: {loc}
  
  ## Impact
  
  High complexity reduces:
  - Code maintainability
  - Test coverage effectiveness
  - Debugging efficiency
  - Onboarding speed for new developers
  
  ## Recommended Actions
  
  1. **Extract Method:** Break down into smaller functions
     - Extract validation logic → `validate{Object}()`
     - Extract error handling → `handle{Error}()`
     - Extract business logic → `process{Operation}()`
  
  2. **Reduce Branching:** Simplify conditional logic
     - Use early returns to reduce nesting
     - Consider strategy pattern for multiple conditions
     - Use lookup tables instead of switch statements
  
  3. **Add Unit Tests:** Target new extracted methods
  
  ## Definition of Done
  
  - [ ] Complexity reduced to ≤ 15
  - [ ] All existing tests still pass
  - [ ] New unit tests added for extracted methods
  - [ ] Code review completed
  
  ## Related Items
  
  - Code Quality Report: #{discussion_number}
  - Similar complexity issues: {related_issues}
  
  ## Priority
  
  {priority_level} - This function is in the {criticality} path.
[/create-issue]
```

**Note:** Create one issue per high-priority finding (up to max 10).

### Step 5: Add Comments to Related Items

If there are open PRs or issues related to the findings, add contextual comments:

```markdown
[add-comment]
target: {related_issue_number}
body: |
  ## 📊 Code Quality Update
  
  This issue was mentioned in the latest code quality analysis.
  
  **Current Status:**
  - Complexity: {current_complexity} (target: ≤ 15)
  - Test Coverage: {current_coverage}% (target: ≥ 80%)
  
  **Related Findings:**
  - New issue created: #{new_issue_number}
  - Full report: #{discussion_number}
  
  **Recommendation:** {specific_recommendation}
[/add-comment]
```

## Key Decision Points

### When to Create Issue vs Include in Discussion Only

| Finding | Create Issue | Discussion Only |
|---------|--------------|-----------------|
| Complexity > 20 | ✅ Yes | Also in discussion |
| Critical lint errors | ✅ Yes | Also in discussion |
| Coverage < 50% for critical modules | ✅ Yes | Also in discussion |
| Medium complexity (15-20) | ❌ No | Discussion only |
| Warnings (not errors) | ❌ No | Discussion only |
| Style issues | ❌ No | Discussion only |

### Prioritization Logic

```javascript
// Prioritize issues for creation (limit to top 10)
const findings = allFindings
  .map(f => ({
    ...f,
    priority: calculatePriority(f)
  }))
  .sort((a, b) => b.priority - a.priority)
  .slice(0, 10);  // Top 10 only

function calculatePriority(finding) {
  let score = 0;
  
  // Complexity contribution
  if (finding.complexity > 30) score += 10;
  else if (finding.complexity > 20) score += 5;
  
  // Criticality of file
  if (finding.isCriticalPath) score += 5;
  
  // Current issues/bugs
  if (finding.hasActiveBugs) score += 3;
  
  // Test coverage
  if (finding.coverage < 50) score += 3;
  
  return score;
}
```

## Success Metrics

- **Comprehensive View:** All findings in one discussion
- **Actionable Tracking:** Top priorities tracked as issues
- **Historical Context:** Week-over-week trends visible
- **Developer Efficiency:** Clear prioritization and recommendations

## Variations

### Pattern 2A: Category-Based Issues

Create issues grouped by category:

```yaml
safe-outputs:
  create-issue:
    max: 5  # One per category
```

Categories:
- Complexity hotspots issue
- Test coverage gaps issue
- Lint errors issue
- Technical debt issue
- Dependency issues issue

### Pattern 2B: Team-Based Routing

Create issues assigned to specific teams:

```markdown
[create-issue]
title: [Frontend Team] Reduce complexity in UI components
assignees: @frontend-team
```

## Related Patterns

- [Conditional Output](./conditional-output.md) - For severity-based routing
- [Fix-or-Report](./fix-or-report.md) - For automated fixes
- [Comment Pattern](./comment-pattern.md) - For update notifications

---

**Pattern Type:** Hierarchical Outputs (Parent-Child)  
**Complexity:** Medium-High  
**Use Cases:** Code quality analysis, audit reports, multi-item findings  
**Related Workflows:** `static-analysis-report.md`, `glossary-maintainer.md`
