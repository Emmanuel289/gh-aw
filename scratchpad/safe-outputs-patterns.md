# Safe Output Patterns - Technical Deep Dive

This document provides implementation details, architecture notes, and advanced patterns for safe outputs in GitHub Agentic Workflows.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Output Type Implementation Details](#output-type-implementation-details)
3. [Multi-Output Coordination](#multi-output-coordination)
4. [Advanced Patterns](#advanced-patterns)
5. [Performance Considerations](#performance-considerations)
6. [Error Handling Strategies](#error-handling-strategies)
7. [Security Implications](#security-implications)
8. [Testing Strategies](#testing-strategies)

## Architecture Overview

### Safe Output Pipeline

```
┌─────────────────┐
│  AI Agent       │
│  Output         │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Parser         │ Extract [create-issue], [add-comment], etc.
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Validator      │ Check permissions, limits, formats
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Orchestrator   │ Sequence jobs, manage dependencies
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Job Generator  │ Create GitHub Actions jobs
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  GitHub API     │ Execute operations
└─────────────────┘
```

### Job Dependency Graph

Multi-output workflows create job dependencies:

```yaml
jobs:
  agent:
    # Main agent execution
  
  create-issue:
    needs: [agent]
    # Creates issues from agent output
  
  create-discussion:
    needs: [agent]
    # Creates discussion from agent output
  
  add-comment:
    needs: [agent, create-issue, create-discussion]
    # Can reference created items via environment variables
```

**Key Dependencies:**
- `add-comment` typically depends on other output jobs (to reference created items)
- `create-pull-request` runs independently (doesn't need issue/discussion numbers)
- `update-issue` can run independently or after `create-issue`

## Output Type Implementation Details

### Create Issue

**Implementation:** `pkg/workflow/safe_outputs.go` → `CreateIssueJob()`

**Key Features:**
- Temporary ID support for cross-referencing
- Label validation against repository labels
- Assignee validation (user or `copilot` special value)
- Expiration via `expires` field

**Temporary ID Pattern:**
```markdown
[create-issue id="vuln-1"]
title: Critical vulnerability
body: See details below
[/create-issue]

[create-issue]
title: Mitigation for {vuln-1}
body: This addresses issue #{vuln-1}
[/create-issue]
```

**Environment Variables:**
```bash
GH_AW_ISSUE_TITLE_PREFIX="[ai] "
GH_AW_ISSUE_LABELS="bug,security"
GH_AW_ISSUE_ALLOWED_LABELS="bug,security,enhancement"
GH_AW_ISSUE_EXPIRES="7"
GH_AW_ASSIGN_COPILOT="true"
GH_AW_TEMPORARY_ID_MAP='{"vuln-1": "123", "vuln-2": "124"}'
```

**Processing Logic:**
1. Parse issue blocks from agent output
2. Validate title, body, labels, assignees
3. Create issue via GitHub API
4. Track temporary ID → real issue number mapping
5. Export mapping for downstream jobs

### Create Discussion

**Implementation:** `pkg/workflow/safe_outputs.go` → `CreateDiscussionJob()`

**Key Features:**
- Category resolution (name, slug, or ID)
- Auto-close older discussions
- GraphQL API usage (required for discussions)

**Category Resolution:**
```javascript
// Supports multiple formats:
category: "general"          // Slug
category: "General"          // Name
category: "DIC_kwDOABCD123" // GraphQL node ID
```

**Close Older Discussions:**
```yaml
safe-outputs:
  create-discussion:
    close-older-discussions: true
```

Automatically closes discussions from the same workflow to keep only the latest.

**GraphQL Query Pattern:**
```graphql
mutation CreateDiscussion {
  createDiscussion(input: {
    repositoryId: $repoId,
    categoryId: $categoryId,
    title: $title,
    body: $body
  }) {
    discussion {
      id
      number
      url
    }
  }
}
```

### Create Pull Request

**Implementation:** `pkg/workflow/safe_outputs.go` → `CreatePullRequestJob()`

**Key Features:**
- File change tracking
- Patch size limits
- Draft PR support
- Cross-repo support

**File Change Processing:**
```markdown
[create-pull-request]
title: Update dependencies
body: Updates npm packages

---file:package.json---
{
  "dependencies": {
    "react": "^18.0.0"
  }
}
---end-file---

---file:package-lock.json---
{...}
---end-file---
[/create-pull-request]
```

**Patch Size Validation:**
```yaml
safe-outputs:
  create-pull-request:
    max-patch-size: 1024  # KB (default)
```

Large patches are rejected to prevent overwhelming reviews.

**Branch Naming:**
```bash
# Pattern: {workflow_id}/{job_name}/{timestamp}
gh-aw/agent/20260131-120000
```

### Add Comment

**Implementation:** `pkg/workflow/safe_outputs.go` → `AddCommentJob()`

**Key Features:**
- Target resolution (triggering, specific number, or wildcard)
- Hide older comments
- Reference created items

**Target Resolution:**
```yaml
target: triggering  # Comment on PR/issue that triggered workflow
target: 123         # Comment on specific issue/PR #123
target: "*"         # Comment on all related items
```

**Environment Variable Injection:**
```bash
GH_AW_CREATED_ISSUE_NUMBER="123"
GH_AW_CREATED_DISCUSSION_NUMBER="456"
GH_AW_CREATED_PULL_REQUEST_NUMBER="789"
```

Available in comment body for cross-referencing:
```markdown
[add-comment]
body: |
  Created issue #${{ env.GH_AW_CREATED_ISSUE_NUMBER }}
  See discussion #${{ env.GH_AW_CREATED_DISCUSSION_NUMBER }}
[/add-comment]
```

### Update Issue

**Implementation:** `pkg/workflow/safe_outputs.go` → `UpdateIssueJob()`

**Key Features:**
- Selective field updates
- Preserve original content option
- Label/assignee modification

**Update Patterns:**
```markdown
[update-issue number="123"]
title: Updated title
labels: [add:bug, remove:needs-triage]
assignees: [add:@user1, remove:@user2]
state: closed
[/update-issue]
```

**Partial Updates:**
- Only specified fields are updated
- Unspecified fields remain unchanged
- Use `append` for body to add content without replacing

## Multi-Output Coordination

### Job Dependencies

**Pattern 1: Sequential (Comment after Issue)**

```yaml
jobs:
  agent:
    ...
  
  create-issue:
    needs: [agent]
    ...
  
  add-comment:
    needs: [agent, create-issue]  # Wait for issue creation
    env:
      GH_AW_CREATED_ISSUE_NUMBER: ${{ needs.create-issue.outputs.issue_number }}
```

**Pattern 2: Parallel (Independent Outputs)**

```yaml
jobs:
  agent:
    ...
  
  create-issue:
    needs: [agent]  # Independent
    ...
  
  create-discussion:
    needs: [agent]  # Independent
    ...
  
  # Both run in parallel
```

**Pattern 3: Fan-out, Fan-in**

```yaml
jobs:
  agent:
    ...
  
  create-issue-1:
    needs: [agent]
  
  create-issue-2:
    needs: [agent]
  
  create-issue-3:
    needs: [agent]
  
  add-comment:
    needs: [agent, create-issue-1, create-issue-2, create-issue-3]
    # Waits for all issues, then comments with summary
```

### Cross-Referencing Patterns

**Pattern A: Issue → Discussion Link**

```markdown
# In create-issue
[create-issue]
body: |
  See full analysis in discussion #{discussion_number}
[/create-issue]

# Requires: create-discussion job runs first
```

**Pattern B: Temporary IDs**

```markdown
[create-issue id="main-issue"]
title: Main issue
[/create-issue]

[create-issue]
title: Sub-task
body: Related to #{main-issue}
[/create-issue]

# System resolves {main-issue} to actual issue number
```

**Pattern C: Environment Variables**

```markdown
[add-comment]
body: |
  Actions taken:
  - Issue: #${{ env.GH_AW_CREATED_ISSUE_NUMBER }}
  - Discussion: #${{ env.GH_AW_CREATED_DISCUSSION_NUMBER }}
[/add-comment]
```

## Advanced Patterns

### Pattern: Conditional Output Generation

**Technique:** Use agent logic to decide which outputs to create

```markdown
# In agent step
if (criticalVulnerabilities.length > 0) {
  // Create issues for critical items
  criticalVulnerabilities.forEach(vuln => {
    output += `[create-issue]
title: ${vuln.title}
body: ${vuln.description}
[/create-issue]\n`;
  });
}

// Always create discussion
output += `[create-discussion]
title: Security Scan Summary
body: ${summary}
[/create-discussion]\n`;
```

**Benefits:**
- Dynamic output based on analysis
- Avoid creating empty issues
- Conditional escalation

### Pattern: Idempotent Operations

**Technique:** Check for existing items before creating

```markdown
# In agent step
const existingIssues = await github.rest.issues.listForRepo({
  owner,
  repo,
  labels: ['automated', 'security'],
  state: 'open'
});

const hasOpenSecurityIssue = existingIssues.data.some(
  issue => issue.title.includes(vulnerabilityName)
);

if (!hasOpenSecurityIssue) {
  // Create new issue
  output += `[create-issue]...`;
} else {
  // Update existing via comment
  output += `[add-comment target="${existingIssues.data[0].number}"]...`;
}
```

**Benefits:**
- Prevents duplicate issues
- Updates existing items instead
- Cleaner issue tracker

### Pattern: Batch Operations with Limits

**Technique:** Create top-N issues, summarize rest in discussion

```markdown
# Prioritize and limit
const topIssues = findings
  .sort((a, b) => b.priority - a.priority)
  .slice(0, 5);  // Max 5 issues

topIssues.forEach(finding => {
  output += `[create-issue]
title: ${finding.title}
body: ${finding.body}
[/create-issue]\n`;
});

// Remaining findings in discussion
const remainingFindings = findings.slice(5);
output += `[create-discussion]
title: Complete Analysis
body: |
  Top issues created: ${topIssues.length}
  
  Additional findings: ${remainingFindings.length}
  ${remainingFindings.map(f => `- ${f.title}`).join('\n')}
[/create-discussion]\n`;
```

**Benefits:**
- Respects `max` limits
- Complete information preserved
- Reduces noise

### Pattern: Progressive Escalation

**Technique:** Track issue persistence across runs

```markdown
# In agent step (using cache or discussion for persistence)
const previousFindings = await loadPreviousFindings();
const currentFindings = await analyzeCurrent();

currentFindings.forEach(finding => {
  const occurrences = countOccurrences(finding, previousFindings);
  
  if (occurrences >= 3) {
    // Third occurrence → create issue
    output += `[create-issue]
title: Persistent issue: ${finding.title}
body: This issue has persisted for ${occurrences} scans.
[/create-issue]\n`;
  } else {
    // Add to discussion only
    discussionBody += `- ${finding.title} (occurrence ${occurrences})\n`;
  }
});
```

**Benefits:**
- Reduces false positives
- Escalates only persistent issues
- Historical tracking

## Performance Considerations

### API Rate Limiting

**GitHub API Limits:**
- Authenticated: 5,000 requests/hour
- Unauthenticated: 60 requests/hour

**Safe Output Consumption:**
- create-issue: 1 request per issue
- create-discussion: 1 GraphQL mutation
- add-comment: 1 request per comment
- create-pull-request: ~3-5 requests (create branch, commit, open PR)

**Optimization Strategies:**

1. **Batch Operations:** Create multiple issues in parallel
2. **Limit Max Values:** Use reasonable `max` limits (5-10, not 100)
3. **Conditional Creation:** Only create outputs when necessary
4. **Cache Results:** Use `tools: cache-memory: true` to persist data

### Job Execution Time

**Typical Durations:**
- create-issue: 5-10 seconds per issue
- create-discussion: 10-15 seconds
- create-pull-request: 30-60 seconds (includes file changes)
- add-comment: 5-10 seconds per comment

**Optimization:**
- Run independent jobs in parallel
- Use job dependencies strategically
- Set appropriate `timeout-minutes`

### Large Payloads

**Limits:**
- Issue/PR body: 65,536 characters
- Discussion body: 65,536 characters
- Comment body: 65,536 characters
- PR patch size: Configurable (default 1024 KB)

**Strategies for Large Content:**
- Use collapsible `<details>` sections
- Link to external artifacts (logs, reports)
- Paginate findings across multiple issues
- Use GitHub Gist for very large content

## Error Handling Strategies

### Validation Errors

**Common Issues:**
- Invalid label (not in repository)
- Invalid assignee (user doesn't exist)
- Invalid category (discussion)
- Malformed markdown

**Handling:**
```yaml
# Strict mode: Fail workflow on validation error
strict: true

# Permissive mode: Log warning, continue
strict: false
```

### API Errors

**Common Issues:**
- Rate limit exceeded
- Insufficient permissions
- Network timeout
- Repository not found

**Handling:**
```javascript
try {
  await createIssue(...);
} catch (error) {
  if (error.status === 403) {
    // Permission denied - log and skip
    console.warn('Permission denied for issue creation');
  } else if (error.status === 429) {
    // Rate limit - wait and retry
    await sleep(60000);
    await createIssue(...);
  } else {
    throw error;  // Fail workflow for unexpected errors
  }
}
```

### Partial Failures

**Scenario:** Some outputs succeed, others fail

**Strategy 1: Best Effort**
```markdown
- Created 3 of 5 issues
- Failed to create discussion (permission denied)
- Created summary comment
```

**Strategy 2: All-or-Nothing**
```markdown
- Validate all outputs first
- Create all or none
- Rollback on any failure
```

**Recommendation:** Use best-effort for production, all-or-nothing for critical workflows

### Recovery Patterns

**Pattern A: Fallback Output**
```markdown
Try:
  create-pull-request
Catch:
  create-issue (with patch attached)
```

**Pattern B: Error Issue**
```markdown
Try:
  analyze and create outputs
Catch:
  create-issue with error details
```

## Security Implications

### Sanitization

**Always sanitize:**
- User-provided content in titles/bodies
- File paths in PR changes
- URLs and external links
- Shell commands in code blocks

**Implementation:**
```javascript
function sanitize(content) {
  return content
    .replace(/[<>]/g, '')  // Remove HTML
    .replace(/`{3}[\s\S]*?`{3}/g, match => {
      // Preserve code blocks but sanitize content
      return match.replace(/\$/g, '\\$');
    });
}
```

### Permission Validation

**Required Permissions:**
```yaml
permissions:
  issues: write       # For create-issue, update-issue
  pull-requests: write  # For create-pull-request, add-comment
  contents: write      # For create-pull-request (file changes)
```

**Validation:**
- Check permissions before job generation
- Fail early with clear error message
- Document required permissions in workflow

### Secrets Handling

**Never include in outputs:**
- API keys
- Passwords
- Private repository paths
- Internal system details
- PII

**Detection:**
```javascript
const SECRET_PATTERNS = [
  /api[_-]?key[_-]?=?['\"]?([a-zA-Z0-9]{32,})/i,
  /password[_-]?=?['\"]?([a-zA-Z0-9]{8,})/i,
  /token[_-]?=?['\"]?([a-zA-Z0-9]{32,})/i
];

function containsSecrets(text) {
  return SECRET_PATTERNS.some(pattern => pattern.test(text));
}
```

### Cross-Repository Safety

**Risks:**
- Creating issues/PRs in unintended repositories
- Leaking internal information to public repos
- Unauthorized access attempts

**Mitigations:**
```yaml
safe-outputs:
  create-issue:
    repository: owner/specific-repo  # Explicit target
  create-pull-request:
    repository: owner/specific-repo
    # Never use ${github.event.repository.name} from untrusted sources
```

## Testing Strategies

### Unit Testing

**Test Safe Output Parser:**
```go
func TestParseSafeOutputs(t *testing.T) {
    input := `[create-issue]
title: Test Issue
body: Test Body
[/create-issue]`
    
    outputs := ParseSafeOutputs(input)
    assert.Equal(t, 1, len(outputs))
    assert.Equal(t, "create-issue", outputs[0].Type)
    assert.Equal(t, "Test Issue", outputs[0].Title)
}
```

**Test Job Generation:**
```go
func TestCreateIssueJob(t *testing.T) {
    config := &SafeOutputConfig{
        CreateIssue: &CreateIssueConfig{
            TitlePrefix: "[test] ",
            Max: 5,
        },
    }
    
    job := CreateIssueJob(config)
    assert.NotNil(t, job)
    assert.Contains(t, job.Env["GH_AW_ISSUE_TITLE_PREFIX"], "[test] ")
}
```

### Integration Testing

**Test Workflow Compilation:**
```bash
# Compile workflow with safe outputs
./gh-aw compile test-workflow.md

# Verify generated jobs
cat .github/workflows/test-workflow.lock.yml | yq '.jobs | keys'
# Expected: agent, create-issue, create-discussion, add-comment
```

**Test Safe Output Execution:**
```bash
# Run workflow in staged mode (preview)
gh workflow run test-workflow.yml -f staged=true

# Verify outputs in step summary (not actual GitHub items)
```

### End-to-End Testing

**Pattern:**
1. Create test repository
2. Run workflow with safe outputs
3. Verify created issues/PRs/discussions
4. Clean up test items

```bash
# Create test issue
gh issue create --title "Test Issue" --body "Test"

# Run workflow
gh workflow run test-workflow.yml

# Wait for completion
gh run watch

# Verify outputs
gh issue list --label "automated"
gh pr list --label "automated"
```

## Related Documentation

- [Safe Outputs System Specification](./safe-outputs-specification.md)
- [Safe Output Messages](./safe-output-messages.md)
- [Safe Output Environment Variables](./safe-output-environment-variables.md)
- [Safe Outputs Guide (User-Facing)](../docs/safe-outputs-guide.md)

---

**Last Updated:** 2026-01-31  
**Target Audience:** Advanced users, contributors, workflow developers  
**Related Issues:** [#12407](https://github.com/githubnext/gh-aw/issues/12407)
