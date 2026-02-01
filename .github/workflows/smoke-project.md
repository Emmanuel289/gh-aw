---
description: Comprehensive smoke test for all project-related features
on: 
  workflow_dispatch:
  schedule: every 24h
  pull_request:
    types: [labeled]
    names: ["smoke-projects"]
permissions:
  contents: read
  pull-requests: read
  issues: read
name: Smoke Test - Projects
engine: copilot
strict: true
project: "https://github.com/orgs/github-agentic-workflows/projects/1"
network:
  allowed:
    - defaults
    - github
tools:
  bash:
    - "*"
  github:
safe-outputs:
  update-project:
    max: 10
    project: "https://github.com/orgs/github-agentic-workflows/projects/1"
  create-project:
    max: 2
    target-owner: "github-agentic-workflows"
    title-prefix: "Smoke Test -"
  copy-project:
    max: 2
    source-project: "https://github.com/orgs/github-agentic-workflows/projects/1"
    target-owner: "github-agentic-workflows"
  create-project-status-update:
    max: 5
    project: "https://github.com/orgs/github-agentic-workflows/projects/1"
  add-comment:
    max: 1
  messages:
    footer: "> 🔬 *Project features validation by [{workflow_name}]({run_url})*"
    run-started: "🔬 Testing all project features... [{workflow_name}]({run_url})"
    run-success: "✅ All project tests passed! [{workflow_name}]({run_url})"
    run-failure: "❌ Project test failed! [{workflow_name}]({run_url}): {status}"
timeout-minutes: 15
---

# Smoke Test: Project Features

**Purpose:** Comprehensive validation of all project-related features including frontmatter configuration and all project safe outputs.

**Test Board:** https://github.com/orgs/github-agentic-workflows/projects/1

**IMPORTANT: Keep outputs concise. Report each test with ✅ or ❌ status.**

## Test Suite Overview

This smoke test validates:
1. **Project Frontmatter** - Automatic safe-outputs configuration from top-level `project:` field
2. **Update Project** - All update-project operations (add, update, archive, views)
3. **Create Project** - Creating new projects with custom views and fields
4. **Copy Project** - Duplicating existing projects
5. **Project Status Updates** - Creating status updates with all status types

---

## Section 1: Project Top-Level Frontmatter

**Purpose:** Validate that the `project:` top-level frontmatter field automatically configures safe outputs.

### Test 1.1: Verify Environment Variables
Check that the `GH_AW_PROJECT_URL` environment variable is set:

```bash
echo "Project URL from env: $GH_AW_PROJECT_URL"
```

Expected: `https://github.com/orgs/github-agentic-workflows/projects/1`

### Test 1.2: Verify Safe Output Configuration
The frontmatter declares the `project:` field. Verify that the compiler automatically created the necessary safe output configurations with defaults:
- `update-project` with max: 100
- `create-project-status-update` with max: 1

### Test 1.3: Test Update Project with Default URL
Create a test draft issue using the default project URL from frontmatter:

```json
{
  "type": "update_project",
  "content_type": "draft_issue",
  "draft_title": "Smoke Test - Frontmatter - Run ${{ github.run_id }}",
  "fields": {
    "status": "Todo"
  }
}
```

### Test 1.4: Test Project Status Update with Default URL
Create a status update using the default project URL:

```json
{
  "type": "create_project_status_update",
  "body": "Smoke test validation - frontmatter project URL working correctly (Run: ${{ github.run_id }})",
  "status": "ON_TRACK"
}
```

---

## Section 2: Update-Project Safe Output

**Purpose:** Validate all operations of the `update-project` safe output.

### Test 2.1: Add Draft Issue
Create a new draft issue in the project:

```json
{
  "type": "update_project",
  "operation": "add_draft_issue",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "draft_title": "Smoke Test Draft - Run ${{ github.run_id }}",
  "draft_body": "This is a test draft issue created by the update-project smoke test.",
  "fields": {
    "status": "Todo"
  }
}
```

### Test 2.2: Update Item Fields
Update the draft issue from Test 2.1 (track the item ID):

```json
{
  "type": "update_project",
  "operation": "update_item",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "item_id": "<item-id-from-test-2.1>",
  "fields": {
    "status": "In Progress"
  }
}
```

### Test 2.3: Archive Item
Archive the draft issue created in Test 2.1:

```json
{
  "type": "update_project",
  "operation": "archive_item",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "item_id": "<item-id-from-test-2.1>"
}
```

---

## Section 3: Create-Project Safe Output

**Purpose:** Validate creating new GitHub Projects V2.

### Test 3.1: Create Basic Project
Create a simple project with minimal configuration:

```json
{
  "type": "create_project",
  "title": "Basic Smoke Test Project",
  "owner": "github-agentic-workflows",
  "description": "Test project created by smoke test - Run ${{ github.run_id }}"
}
```

### Test 3.2: Create Project with Custom Views
Create a project with multiple views:

```json
{
  "type": "create_project",
  "title": "Multi-View Test Project",
  "owner": "github-agentic-workflows",
  "description": "Test project with custom views - Run ${{ github.run_id }}",
  "views": [
    {
      "name": "Backlog",
      "layout": "TABLE"
    },
    {
      "name": "Current Sprint",
      "layout": "BOARD"
    }
  ]
}
```

### Test 3.3: Verify Max Limit
The frontmatter configures `max: 2` for create-project. After creating 2 projects, verify that a 3rd attempt is blocked by the max limit.

---

## Section 4: Copy-Project Safe Output

**Purpose:** Validate duplicating existing projects.

### Test 4.1: Basic Project Copy
Copy the source project without draft issues:

```json
{
  "type": "copy_project",
  "source_project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "target_owner": "github-agentic-workflows",
  "new_title": "Smoke Test Copy 1 - Run ${{ github.run_id }}",
  "include_draft_issues": false
}
```

### Test 4.2: Copy with Draft Issues
Copy the source project including draft issues:

```json
{
  "type": "copy_project",
  "source_project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "target_owner": "github-agentic-workflows",
  "new_title": "Smoke Test Copy 2 - Run ${{ github.run_id }}",
  "include_draft_issues": true
}
```

### Test 4.3: Verify Max Limit
The frontmatter configures `max: 2` for copy-project. Verify that a 3rd copy attempt is blocked by the max limit.

---

## Section 5: Create-Project-Status-Update Safe Output

**Purpose:** Validate creating project status updates.

### Test 5.1: ON_TRACK Status Update
Create a status update with ON_TRACK status:

```json
{
  "type": "create_project_status_update",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "status": "ON_TRACK",
  "body": "✅ Smoke test validation in progress - Run ${{ github.run_id }}. All systems nominal."
}
```

### Test 5.2: AT_RISK Status Update
Create a status update with AT_RISK status:

```json
{
  "type": "create_project_status_update",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "status": "AT_RISK",
  "body": "⚠️ Testing AT_RISK status - Run ${{ github.run_id }}. This is a simulated at-risk condition for testing purposes."
}
```

### Test 5.3: OFF_TRACK Status Update
Create a status update with OFF_TRACK status:

```json
{
  "type": "create_project_status_update",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "status": "OFF_TRACK",
  "body": "🔴 Testing OFF_TRACK status - Run ${{ github.run_id }}. This is a simulated off-track condition for testing purposes."
}
```

### Test 5.4: Status Update with Markdown
Create a status update with rich markdown formatting:

```json
{
  "type": "create_project_status_update",
  "project": "https://github.com/orgs/github-agentic-workflows/projects/1",
  "status": "ON_TRACK",
  "body": "## Smoke Test Progress Report\n\n**Run ID:** ${{ github.run_id }}\n\n### Completed:\n- ✅ Frontmatter tests\n- ✅ Update-project tests\n- ✅ Create-project tests\n- ✅ Copy-project tests\n\n### In Progress:\n- 🔄 Status update tests\n\n> 💡 **Note:** This is an automated smoke test validation."
}
```

### Test 5.5: Verify Default Project URL
Create a status update without specifying the project (should use frontmatter default):

```json
{
  "type": "create_project_status_update",
  "status": "ON_TRACK",
  "body": "Testing default project URL from frontmatter - Run ${{ github.run_id }}"
}
```

---

## Output Requirements

Add a **concise comment** to the pull request (if triggered by PR) with test results:

### Section 1: Project Frontmatter
| Test | Status | Notes |
|------|--------|-------|
| Environment variable set | ✅/❌ | GH_AW_PROJECT_URL |
| Auto-configured update-project | ✅/❌ | Max: 100 |
| Auto-configured status updates | ✅/❌ | Max: 1 |
| Update project with default URL | ✅/❌ | Draft created |
| Status update with default URL | ✅/❌ | Status posted |

### Section 2: Update-Project
| Operation | Status | Notes |
|-----------|--------|-------|
| Add draft issue | ✅/❌ | Item ID created |
| Update item fields | ✅/❌ | Status changed |
| Archive item | ✅/❌ | Item archived |

### Section 3: Create-Project
| Test Case | Status | Project URL |
|-----------|--------|-------------|
| Basic project | ✅/❌ | [URL] |
| Multi-view project | ✅/❌ | [URL] |
| Max limit enforced | ✅/❌ | 3rd blocked |

### Section 4: Copy-Project
| Test Case | Status | Project URL |
|-----------|--------|-------------|
| Copy without items | ✅/❌ | [URL] |
| Copy with items | ✅/❌ | [URL] |
| Max limit enforced | ✅/❌ | 3rd blocked |

### Section 5: Status Updates
| Test Case | Status | Notes |
|-----------|--------|-------|
| ON_TRACK status | ✅/❌ | Posted |
| AT_RISK status | ✅/❌ | Posted |
| OFF_TRACK status | ✅/❌ | Posted |
| Markdown formatting | ✅/❌ | Rendered |
| Default project URL | ✅/❌ | Used default |

**Overall Status:** PASS / FAIL

---

## Success Criteria

- All frontmatter tests pass (environment variables, auto-configuration)
- All update-project operations succeed
- Projects are created with custom views
- Projects are copied correctly with/without items
- All status update types work (ON_TRACK, AT_RISK, OFF_TRACK)
- Markdown formatting renders properly in status updates
- Max limits are enforced for all safe outputs
- No errors in workflow logs

## Cleanup

After the test completes:
1. Manually archive or delete the created test projects to avoid clutter
2. Archive draft issues created during testing
3. Verify all test artifacts are cleaned up from the test board
