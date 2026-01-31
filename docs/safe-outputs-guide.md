# Safe Output Patterns and Best Practices

This guide helps you choose the right safe output type and design effective multi-output workflows for GitHub Agentic Workflows.

## Quick Reference

**Safe Output Types**:
- `create-issue` - Action required, task tracking, assignment
- `create-discussion` - Analysis, reports, community conversation
- `create-pull-request` - Automated fixes, code changes
- `add-comment` - Updates to existing items, progress reports
- `update-issue` - Modify existing issue content

## When to Use Each Output Type

### Create Issue (`create-issue`)

**Use when:**
- ✅ Action is required (task tracking)
- ✅ Assignment to developer/agent needed
- ✅ Status tracking required (open/closed)
- ✅ Needs labels, milestones, or project boards
- ✅ Workflow needs to create sub-tasks
- ✅ Time-sensitive action items need visibility

**Examples:**
- Security vulnerability found in dependencies
- Test failures requiring fixes
- Code quality issues needing remediation
- Breaking changes detected in PR
- Expired certificates or credentials
- Resource limits exceeded

**Typical Configuration:**
```yaml
safe-outputs:
  create-issue:
    title-prefix: "[automated] "
    labels: [automation, needs-triage]
    assignees: copilot
    expires: 7d  # Auto-close if condition resolves
    max: 5       # Limit number of issues per run
```

**Best Practices:**
- Use `expires` for transient issues (security scans, CI failures)
- Add descriptive labels for filtering and triage
- Include clear action items in issue body
- Link to source discussion or PR for context

### Create Discussion (`create-discussion`)

**Use when:**
- ✅ Sharing analysis, reports, or insights
- ✅ No specific action required
- ✅ Community conversation welcome
- ✅ Long-term reference material needed
- ✅ Regular status reports or summaries
- ✅ Archive-worthy findings

**Examples:**
- Performance analysis reports
- Security scan summaries
- Trend analysis over time
- Weekly/monthly statistics
- Code quality metrics dashboard
- Audit logs and compliance reports

**Typical Configuration:**
```yaml
safe-outputs:
  create-discussion:
    category: "Reports"  # or "Audits", "Security", etc.
    title-prefix: "[Report] "
    close-older-discussions: true  # Keep only latest
    max: 1  # One discussion per run
```

**Best Practices:**
- Use `close-older-discussions: true` for time-sensitive reports
- Choose appropriate category (Reports, Audits, Security)
- Structure content with clear sections (Summary, Details, Action Items)
- Link to related issues for actionable items
- Include date/timestamp in title for historical tracking

### Create Pull Request (`create-pull-request`)

**Use when:**
- ✅ Automated fix can be proposed
- ✅ Code changes are ready to review
- ✅ Changes are non-breaking and safe
- ✅ Tests can validate the changes
- ✅ Human review is needed before merge

**Examples:**
- Dependency updates
- Code formatting fixes
- Documentation updates
- Configuration file corrections
- License header updates
- Automated refactoring

**Typical Configuration:**
```yaml
safe-outputs:
  create-pull-request:
    title-prefix: "[auto-fix] "
    labels: [automation, safe-to-merge]
    draft: true  # Create as draft PR (default)
    expires: 14d  # Auto-close stale PRs
    if-no-changes: warn  # warn, error, or ignore
```

**Best Practices:**
- Create as draft PR by default (safer)
- Include detailed description of changes and motivation
- Link to triggering issue or discussion
- Add tests or validation steps in PR description
- Use `expires` to prevent stale PR accumulation

### Add Comment (`add-comment`)

**Use when:**
- ✅ Updating existing issue/PR/discussion
- ✅ Providing progress updates
- ✅ Responding to triggers
- ✅ Summarizing results of analysis
- ✅ Linking related items

**Examples:**
- CI results on PR
- Approval status updates
- Analysis updates on issues
- Progress reports on long-running tasks
- Links to created issues/discussions/PRs

**Typical Configuration:**
```yaml
safe-outputs:
  add-comment:
    target: triggering  # or specific number, or "*" for all
    hide-older-comments: true  # Hide previous comments
    max: 1  # Limit comments per run
```

**Best Practices:**
- Use `target: triggering` to comment on source item
- Use `hide-older-comments: true` to reduce noise
- Keep comments concise and actionable
- Link to detailed discussion/issue for full analysis
- Include status indicators (✅, ⚠️, ❌) for quick scanning

### Update Issue (`update-issue`)

**Use when:**
- ✅ Need to modify existing issue content
- ✅ Updating status or assignees
- ✅ Adding/removing labels
- ✅ Changing milestones or project board

**Examples:**
- Updating issue body with analysis results
- Changing assignees based on triage
- Adding labels after classification
- Updating issue title for clarity

**Typical Configuration:**
```yaml
safe-outputs:
  update-issue:
    max: 10  # Limit updates per run
```

**Best Practices:**
- Preserve original issue content when possible
- Add comments explaining what changed and why
- Use sparingly - prefer comments for updates
- Ensure idempotency (safe to run multiple times)

## Multi-Output Workflow Patterns

74.8% of agentic workflows use multiple safe output types. Here are proven patterns:

### Pattern 1: Conditional Outputs (Decision-Based)

**When to use:** Output type depends on analysis results.

**Example - Security Scanning:**
```yaml
safe-outputs:
  create-issue:
    title-prefix: "[security] "
    labels: [security, critical]
    max: 5
  create-discussion:
    category: "Security"
    title-prefix: "[Security Scan] "
    max: 1
```

**Logic:**
- Critical findings (CVSS ≥ 7.0) → Create issue per vulnerability
- Low/medium findings → Create discussion with summary
- Always → Add comment to PR with scan results

**Benefits:**
- Critical items get immediate visibility
- Non-critical items archived for reference
- Reduces noise from low-severity findings

See [conditional-output.md](./examples/safe-outputs/conditional-output.md) for complete example.

### Pattern 2: Hierarchical Outputs (Parent-Child)

**When to use:** Create summary with detailed sub-items.

**Example - Code Quality Analysis:**
```yaml
safe-outputs:
  create-discussion:
    category: "Reports"
    title-prefix: "[Code Quality] "
    max: 1
  create-issue:
    title-prefix: "[quality] "
    labels: [code-quality, automated]
    max: 10
  add-comment:
    target: triggering
    max: 1
```

**Logic:**
- Create discussion with overall analysis
- Create issues for each actionable finding
- Add comment linking to discussion and issues

**Benefits:**
- Comprehensive view in discussion
- Individual tracking per issue
- Easy to assign and close items independently

See [multi-output-analysis.md](./examples/safe-outputs/multi-output-analysis.md) for complete example.

### Pattern 3: Fix-or-Report (Progressive)

**When to use:** Attempt automated fix, report if unable.

**Example - Dependency Updates:**
```yaml
safe-outputs:
  create-pull-request:
    title-prefix: "[deps] "
    labels: [dependencies]
    max: 1
  create-issue:
    title-prefix: "[deps-failed] "
    labels: [dependencies, needs-review]
    max: 1
  add-comment:
    target: triggering
    max: 1
```

**Logic:**
- If update is safe → Create PR with changes
- If conflicts/issues → Create issue for manual review
- Always → Add comment with summary and links

**Benefits:**
- Automates safe updates
- Surfaces problematic updates for human review
- Provides complete audit trail

See [fix-or-report.md](./examples/safe-outputs/fix-or-report.md) for complete example.

### Pattern 4: Comment-First (Update-Focused)

**When to use:** Primarily updating existing items.

**Example - CI Status Reporter:**
```yaml
safe-outputs:
  add-comment:
    target: triggering
    hide-older-comments: true
    max: 1
  create-issue:
    title-prefix: "[ci-failure] "
    labels: [ci, bug]
    max: 1  # Only if persistent failure
```

**Logic:**
- Always → Add comment with CI results
- If persistent failure (3+ runs) → Create issue
- Hide older comments to reduce clutter

**Benefits:**
- Keeps all status in one place
- Escalates only persistent problems
- Clean comment history

See [comment-pattern.md](./examples/safe-outputs/comment-pattern.md) for complete example.

## Best Practices

### Output Hygiene

**Avoid Duplication:**
```yaml
# ❌ BAD - Creates both issue and discussion for same finding
safe-outputs:
  create-issue:
    max: 10
  create-discussion:
    max: 1
# Agent creates issue and discussion for each vulnerability
```

```yaml
# ✅ GOOD - Conditional based on severity
safe-outputs:
  create-issue:
    max: 5  # Only critical
  create-discussion:
    max: 1  # Summary of all
# Agent creates issues only for critical, discussion for summary
```

**Cleanup Transient Items:**
```yaml
# ✅ Use expires for time-sensitive items
safe-outputs:
  create-issue:
    expires: 7d  # Auto-close when no longer relevant
  create-discussion:
    close-older-discussions: true  # Keep only latest report
```

**Limit Output Volume:**
```yaml
# ✅ Set reasonable max values
safe-outputs:
  create-issue:
    max: 5  # Don't overwhelm with 100 issues
  add-comment:
    max: 1  # One summary comment, not per-item
```

### User Experience

**Consistent Output Types:**
- Security workflows → Issues (critical) + Discussion (summary)
- Performance reports → Discussion (always)
- Automated fixes → Pull Request (primary)
- Status updates → Comment (always)

**Clear Titles and Prefixes:**
```yaml
# ✅ GOOD - Clear, scannable titles
safe-outputs:
  create-issue:
    title-prefix: "[security-critical] "
  create-discussion:
    title-prefix: "[Security Scan] "
```

**Standard Body Structure:**
```markdown
## Summary
Brief overview of findings

## Details
<details>
<summary>Detailed Analysis</summary>
Full technical details
</details>

## Action Items
- [ ] Task 1
- [ ] Task 2

## Related Items
- Discussion: #123
- Issue: #456
```

**Attribution and Context:**
- Always include workflow run link (automatic)
- Link to triggering item when relevant
- Reference related outputs (issue → discussion)

### Error Handling

**Always Produce Output:**
```yaml
# Even on workflow failure, create output
safe-outputs:
  create-discussion:  # Fallback for errors
    category: "Reports"
    max: 1
```

**Don't Create Issues for Workflow Failures:**
```yaml
# ❌ BAD - Creates noise
safe-outputs:
  create-issue:
    title-prefix: "[workflow-error] "
# Don't create issues when the workflow itself fails

# ✅ GOOD - Use discussion for workflow issues
safe-outputs:
  create-discussion:
    category: "Audits"
    title-prefix: "[Workflow Status] "
```

**Include Debug Information:**
- Workflow run URL (automatic)
- Relevant environment variables
- Error messages and stack traces
- Steps to reproduce

### Security Considerations

**Sanitize Sensitive Data:**
```yaml
# ⚠️ Be careful with safe outputs containing:
# - API keys or tokens
# - Private repository paths
# - Internal system details
# - PII or confidential information
```

**Use Appropriate Visibility:**
- Public repos → All outputs are public
- Private repos → Outputs inherit repo visibility
- Cross-repo → Ensure target repo permissions are correct

**Validate Input:**
- Sanitize user-provided content
- Validate issue/PR numbers before referencing
- Check permissions before creating cross-repo outputs

## Decision Tree

Use this flowchart to choose the right output type:

```
Start: What is the goal of the output?

├─ Is action required?
│  ├─ Yes: Does it need assignment/tracking?
│  │  ├─ Yes → create-issue
│  │  └─ No: Can it be automated?
│  │     ├─ Yes → create-pull-request
│  │     └─ No → create-issue
│  └─ No: Is it updating existing item?
│     ├─ Yes → add-comment
│     └─ No: Is it a report/analysis?
│        ├─ Yes → create-discussion
│        └─ No → create-discussion or add-comment
```

**Quick Decision Table:**

| Scenario | Primary Output | Secondary Output | Notes |
|----------|---------------|------------------|-------|
| Critical vulnerability | `create-issue` | `add-comment` | Issue for tracking, comment for context |
| Security scan summary | `create-discussion` | `create-issue` (critical only) | Discussion for all, issues for urgent |
| Dependency update | `create-pull-request` | `create-issue` (if fails) | PR first, issue as fallback |
| CI results | `add-comment` | `create-issue` (persistent failure) | Comment for status, escalate failures |
| Performance report | `create-discussion` | - | Discussion only, no action needed |
| Code quality fix | `create-pull-request` | `add-comment` | PR for fix, comment for summary |
| Triage results | `add-comment` + `add-labels` | `create-discussion` (summary) | Comment per issue, discussion for batch |

## Examples

Complete example workflows demonstrating these patterns:

- [Conditional Output](./examples/safe-outputs/conditional-output.md) - Dynamic output type selection based on severity
- [Multi-Output Analysis](./examples/safe-outputs/multi-output-analysis.md) - Discussion with sub-issues pattern
- [Fix-or-Report](./examples/safe-outputs/fix-or-report.md) - Attempt PR, fallback to issue
- [Comment Pattern](./examples/safe-outputs/comment-pattern.md) - Status updates with escalation

## Related Documentation

- [Safe Outputs System Specification](../scratchpad/safe-outputs-specification.md) - Formal specification
- [Safe Output Messages](../scratchpad/safe-output-messages.md) - Message templates and formatting
- [Safe Output Environment Variables](../scratchpad/safe-output-environment-variables.md) - Configuration reference
- [Safe Output Patterns (Technical)](../scratchpad/safe-outputs-patterns.md) - Implementation details

## Statistics

Based on analysis of 147 agentic workflows (as of 2026-01):

| Output Type | Usage | Percentage |
|-------------|-------|------------|
| `create-issue` | 113 workflows | 76.9% |
| `create-discussion` | 108 workflows | 73.5% |
| `create-pull-request` | 54 workflows | 36.7% |
| `add-comment` | 34 workflows | 23.1% |
| `update-issue` | 4 workflows | 2.7% |

**Multi-Output Workflows:** 110 workflows (74.8%) use 2+ output types

**Key Insights:**
- Most workflows combine issues and discussions (comprehensive reporting)
- PRs are less common but critical for automation
- Comments are typically used in conjunction with other outputs
- Update operations are rare (prefer comments for updates)

---

**Last Updated:** 2026-01-31  
**Related Issues:** [#12407](https://github.com/githubnext/gh-aw/issues/12407)
