# ProjectOps + Orchestration: Analysis and Recommendations

## Executive Summary

This document analyzes the current pain points in combining projectOps with orchestration patterns in GitHub Agentic Workflows and provides minimal, architectural recommendations that naturally fit into the existing architecture.

## Current Architecture

### Orchestration Pattern
- **Orchestrator**: Dispatches workers and aggregators using `dispatch-workflow`
- **Workers**: Process individual units of work (e.g., triage sub-issues)
- **Aggregator**: Collects results and updates tracking issues/projects

### ProjectOps Pattern
- Safe outputs: `create-issue`, `link-sub-issue`, `update-issue`, `update-project`
- Temporary ID system for referencing newly created issues before they exist
- Island-based content updates using `replace-island` operation (currently run-id based only)

## Pain Points Analysis

### 1. Tool-Call Ordering Flakiness

**Problem**: Agent sometimes tries to link a sub-issue before creating it, breaking temporary ID references.

**Root Cause**: 
- LLM non-determinism in tool call sequencing
- Lack of explicit dependency enforcement in the tool system

**Current Workaround**: 
- Stricter prompt structure (STEP 1/STEP 2)
- Explicit "DO NOT" constraints
- Results: 90% reliability but still fragile

**Recommended Solution**: **Safe Output Dependency Chains** (Minimal Addition)

Add optional `depends_on` field to safe output messages that defers execution until dependencies resolve:

```javascript
// Agent output with dependency
{
  "type": "link_sub_issue",
  "parent_issue_number": 123,
  "sub_issue_number": "aw_temp_001",
  "depends_on": ["aw_temp_001"]  // Defers until this temporary ID resolves
}
```

**Implementation**:
- Modify `actions/setup/js/handler_manager.cjs` to track unresolved dependencies
- Add dependency resolution phase before processing safe output messages
- Already partially implemented: `link_sub_issue.cjs` has deferred status for unresolved temp IDs (lines 70-88)

**Impact**: Low complexity, natural extension of existing temporary ID system

---

### 2. Missing Optional Parameters (island_id)

**Problem**: `island_id` missing on `update_issue` causes aggregator to duplicate compliance sections instead of replacing them.

**Root Cause**:
- Current `replace-island` operation uses `runId` as the island identifier (automatic)
- No way to specify a **named island** for deterministic, cross-run updates
- Aggregator can't update the same section across multiple runs

**Current Workaround**:
- Moved to deterministic flow: read → find section → remove → rewrite
- Fragile and prone to formatting errors

**Recommended Solution**: **Named Islands** (Minimal Addition)

Extend `replace-island` operation to support optional `island_id` parameter:

```yaml
# Frontmatter
safe-outputs:
  update-issue:
    body: true
    operation: replace-island  # Allow named islands
    max: 5
```

```javascript
// Agent output with named island
{
  "type": "update_issue",
  "issue_number": 123,
  "operation": "replace-island",
  "island_id": "compliance-summary",  // Named island (optional)
  "body": "## Compliance Status\n\n- Worker 1: ✅ Complete\n- Worker 2: ✅ Complete"
}
```

**Implementation Changes**:

1. **Schema Update** (`pkg/parser/schemas/main_workflow_schema.json`):
   - Add `island_id` as optional string field in update-issue tool

2. **JavaScript Update** (`actions/setup/js/update_pr_description_helpers.cjs`):
   ```javascript
   // Extend buildIslandStartMarker to support named islands
   function buildIslandStartMarker(runId, islandId) {
     if (islandId) {
       return `<!-- gh-aw-island-start:${islandId} -->`;
     }
     return `<!-- gh-aw-island-start:${runId} -->`;
   }
   ```

3. **Update Handler** (`actions/setup/js/update_issue.cjs`):
   - Pass `island_id` from message to `updateBody()` function
   - Default to `runId` if `island_id` not provided (backward compatible)

**Benefits**:
- Deterministic updates: Same island updated across multiple workflow runs
- Aggregator can reliably update specific sections without duplicating
- Backward compatible: Falls back to `runId` if `island_id` not specified
- Minimal changes to existing code

**Impact**: Low complexity, natural extension of existing replace-island feature

---

### 3. Orchestration Timing Dependencies

**Problem**: Aggregator runs before workers finish, reports everything as "Pending"

**Root Cause**:
- `dispatch-workflow` launches workers asynchronously
- Aggregator starts in parallel, doesn't wait for workers
- No built-in synchronization mechanism

**Current Workaround**:
- Polling/retry in aggregator (fragile)
- 90-second delays (still not enough)
- Timing-dependent and unreliable

**Recommended Solution A**: **Workflow Completion Events** (Requires GitHub Actions Enhancement)

**NOT RECOMMENDED** - Requires GitHub Actions to support `workflow_run` events for `workflow_dispatch` triggers, which is not currently supported.

**Recommended Solution B**: **Polling with Project Status Fields** (Minimal, Uses Existing Infrastructure)

Use GitHub Projects as a coordination mechanism:

```yaml
# Worker workflow frontmatter
safe-outputs:
  update-project:
    project: "https://github.com/orgs/github/projects/24060"
    max: 1
```

```javascript
// Worker: Update project status when complete
update_project({
  project: "...",
  content_type: "draft_issue",
  draft_title: "Worker Status",
  fields: {
    "Status": "Complete",  // Workers mark themselves complete
    "Worker ID": "worker-1",
    "Completed At": "2026-02-07T06:00:00Z"
  }
})
```

```javascript
// Aggregator: Poll project items until all workers complete
const allWorkersComplete = await checkProjectItems({
  project: "...",
  filter: { "Status": "Complete" },
  expectedCount: 5  // Number of workers dispatched
});

if (!allWorkersComplete) {
  // Defer aggregation or update status as "In Progress"
  return;
}

// All workers done, proceed with aggregation
```

**Implementation**:
- No new safe outputs needed - uses existing `update-project`
- Aggregator includes polling logic to check project status fields
- Workers update project status when complete

**Benefits**:
- Uses existing GitHub Projects infrastructure
- Natural status tracking for monitoring
- No new safe outputs or features required
- Self-documenting: Project board shows worker progress

**Alternative**: **Wait-for-Completion Safe Output** (New Feature, More Reliable)

Add new `wait-for-workflows` safe output that blocks until dispatched workflows complete:

```yaml
safe-outputs:
  dispatch-workflow:
    workflows: [worker-a, worker-b]
    max: 10
  
  wait-for-workflows:
    timeout: 300  # 5 minutes
    poll-interval: 15  # Check every 15 seconds
```

```javascript
// Orchestrator dispatches workers and waits
worker_a({ tracker_id: 123 });
worker_b({ tracker_id: 123 });

// Wait for all dispatched workflows to complete
wait_for_workflows({
  timeout: 300  // 5 minutes max wait
});

// Now safe to aggregate
```

**Implementation Complexity**: Medium
- Requires tracking dispatched workflow run IDs
- Needs GitHub API polling for workflow status
- Timeout and error handling

**Impact**: Medium complexity, but provides reliable synchronization

**Recommendation**: Start with **Solution B (Project Status Polling)** as it requires no new features and is immediately implementable. Consider **wait-for-workflows** if Project polling proves insufficient.

---

### 4. Event Trigger Re-entrancy / Cascading Runs

**Problem**: Creating and labeling sub-issues triggers dispatcher again via `issues.labeled`, causing loops and unnecessary compute.

**Root Cause**:
- Workflows triggered by `issues.labeled` or `issues.opened`
- Sub-issue creation/labeling creates events that re-trigger the dispatcher
- No built-in filtering to distinguish orchestrator-created issues from user-created issues

**Current Workaround**:
- Filtering at GitHub Actions level (`if` conditions)
- Checking for specific labels or markers in workflow YAML
- Requires manual configuration in each workflow

**Recommended Solution**: **Built-in Re-entrancy Protection** (Safe Output Enhancement)

Add automatic re-entrancy markers to created issues:

```yaml
# Frontmatter - enable re-entrancy protection
safe-outputs:
  create-issue:
    max: 10
    labels: [task]
    prevent-retrigger: true  # Add marker to prevent re-triggering
```

**Implementation**:

1. **Automatic Marker Addition** (`actions/setup/js/create_issue.cjs`):
   ```javascript
   // Add hidden HTML comment marker to issue body
   const reentrantMarker = `<!-- gh-aw-created-by:${workflowName}:${runId} -->`;
   const bodyWithMarker = reentrantMarker + "\n\n" + issueBody;
   ```

2. **Workflow Conditional Generation** (`pkg/workflow/compiler_yaml.go`):
   ```yaml
   # Generated workflow includes automatic re-entrancy check
   on:
     issues:
       types: [opened, labeled]
   
   jobs:
     agent:
       if: |
         !contains(github.event.issue.body, '<!-- gh-aw-created-by:')
   ```

3. **Schema Addition** (`pkg/parser/schemas/main_workflow_schema.json`):
   - Add `prevent-retrigger` boolean field to create-issue configuration

**Benefits**:
- Automatic protection without manual workflow configuration
- Invisible to users (hidden HTML comments)
- Works across all orchestration patterns
- Backward compatible (opt-in via frontmatter)

**Alternative**: **Explicit Creation Labels**

Add special label like `gh-aw-orchestrated` to all orchestrator-created issues:

```yaml
safe-outputs:
  create-issue:
    labels: [task, gh-aw-orchestrated]  # Always include marker label
```

Then filter in dispatcher workflow:

```yaml
on:
  issues:
    types: [opened, labeled]

jobs:
  agent:
    if: |
      !contains(github.event.issue.labels.*.name, 'gh-aw-orchestrated')
```

**Implementation**: Simpler but requires user configuration

**Recommendation**: **Built-in re-entrancy protection** is preferred as it's automatic and doesn't pollute the label namespace.

**Impact**: Low complexity, natural extension of existing safe outputs

---

## Summary of Recommendations

### Immediate (Low Effort, High Impact)

1. **Named Islands** (`island_id` parameter)
   - Extend `replace-island` to support named islands
   - Implementation: ~50 LOC changes to `update_pr_description_helpers.cjs` and schema
   - Impact: Solves aggregator duplication problem completely
   - **Priority: HIGH**

2. **Re-entrancy Protection** (`prevent-retrigger` flag)
   - Add automatic markers to created issues to prevent re-triggering
   - Implementation: ~100 LOC changes to `create_issue.cjs` and compiler
   - Impact: Eliminates cascading runs and unnecessary compute
   - **Priority: HIGH**

### Near-term (Medium Effort, High Impact)

3. **Project Status Polling Pattern** (Documentation + Example)
   - Document how to use existing `update-project` for coordination
   - Provide example orchestrator/aggregator workflows
   - Implementation: Documentation only, no code changes
   - Impact: Solves timing dependency without new features
   - **Priority: MEDIUM**

### Future (Medium Effort, High Value)

4. **Dependency Chains** (`depends_on` field)
   - Formalize existing deferred execution for temporary IDs
   - Implementation: ~150 LOC changes to `handler_manager.cjs`
   - Impact: Reduces prompt engineering burden, more reliable sequencing
   - **Priority: MEDIUM** (current workarounds are 90% effective)

5. **Wait-for-Workflows** (New Safe Output)
   - Add explicit synchronization primitive for workflow completion
   - Implementation: ~300 LOC new safe output handler
   - Impact: Most reliable solution for timing dependencies
   - **Priority: LOW** (Project polling pattern sufficient for most cases)

---

## Implementation Priority

```
High Priority (Implement First):
├── Named Islands (island_id)
└── Re-entrancy Protection (prevent-retrigger)

Medium Priority (Implement After High):
├── Project Status Polling (Documentation)
└── Dependency Chains (depends_on)

Low Priority (Evaluate After Medium):
└── Wait-for-Workflows (New Safe Output)
```

---

## Key Principles

All recommendations follow these principles:

1. **Minimal Changes**: Extend existing features rather than adding new ones
2. **Natural Fit**: Align with current architecture and patterns
3. **Backward Compatible**: Don't break existing workflows
4. **Self-Documenting**: Use clear, intuitive names and patterns
5. **No Workarounds**: Solve root cause, not symptoms

---

## Conclusion

The pain points in combining projectOps with orchestration are real but solvable with **minimal, targeted additions** to the existing architecture:

- **Named Islands** solves parameter omission and aggregator duplication
- **Re-entrancy Protection** eliminates cascading runs automatically  
- **Project Status Polling** provides timing coordination without new features
- **Dependency Chains** and **Wait-for-Workflows** are nice-to-have enhancements for the future

These recommendations preserve the existing architecture while addressing the specific pain points identified. The highest-priority items (**Named Islands** and **Re-entrancy Protection**) can be implemented with ~150 LOC combined and provide immediate, significant value.
