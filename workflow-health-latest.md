# Workflow Health Dashboard - 2026-02-10

## Overview
- **Total workflows**: 148 (148 executable, 59 shared includes)
- **Healthy**: 137 (92.6%)
- **Warning**: 11 (7.4%)
- **Critical**: 0 (0%)
- **Inactive**: 0 (0%)
- **Compilation coverage**: 148/148 (100% ✅)
- **Overall health score**: 78/100 (↓ -19 from 97/100)

## 🟡 STATUS: WARNING - Minor Issues Detected

### Health Assessment Summary

**Status: WARNING (Declined from EXCELLENT)**

The ecosystem has **11 outdated lock files** and **1 failing workflow**:
- ⚠️ **11 workflows with outdated lock files** (need recompilation)
- ❌ **1 workflow failing** (daily-fact - missing JavaScript module)
- ✅ **100% compilation coverage** maintained (all 148 workflows have locks)
- ✅ **137 healthy workflows** (92.6%)
- ⚠️ **Health score dropped by -19 points** (97 → 78)

**Key Changes Since Last Check (2026-02-09):**
- ↓ Health score decreased by -19 points (97 → 78)
- ⚠️ 11 workflows need recompilation (source .md newer than .lock.yml)
- ❌ daily-fact workflow failing (missing handle_noop_message.cjs)
- ✅ Compilation coverage still at 100% (148/148)

## Critical Issues 🚨

### daily-fact Workflow (Priority: P1)

**Status:** Failing due to missing JavaScript module

**Error Details:**
- **Run**: [§21862815504](https://github.com/github/gh-aw/actions/runs/21862815504)
- **Failed Job**: conclusion (ID: 63096273090)
- **Error**: `Cannot find module '/opt/gh-aw/actions/handle_noop_message.cjs'`
- **Impact**: Workflow cannot complete - conclusion job fails every time
- **Engine**: codex (id: codex)
- **Description**: Posts a daily poetic verse about gh-aw to a discussion thread

**Root Cause:**
The workflow's conclusion job tries to load `handle_noop_message.cjs` which doesn't exist in `/opt/gh-aw/actions/`. This appears to be a missing file in the actions setup that handles noop safe outputs.

**Recommended Fix:**
1. Check if `handle_noop_message.cjs` should exist in `actions/setup/js/`
2. If missing, create the module or update workflow to not require it
3. If file exists but not deployed, update `actions/setup/setup.sh` to include it
4. Recompile workflow after fix: `gh aw compile .github/workflows/daily-fact.md`

**Workaround:**
The workflow created issue [#14763](https://github.com/github/gh-aw/issues/14763) successfully before failing, so partial functionality is preserved.

## Warnings ⚠️

### Outdated Lock Files (11 workflows)

The following workflows have source `.md` files that are newer than their compiled `.lock.yml` files. They need recompilation to pick up latest changes:

<details>
<summary><b>View Outdated Workflows (11 total)</b></summary>

1. **auto-triage-issues** - Automatically labels new and existing unlabeled issues
   - Engine: copilot
   - Has tools: ✅ | Safe outputs: ✅

2. **daily-code-metrics** - Code metrics analysis
   - Has tools: ✅ | Safe outputs: ✅

3. **daily-observability-report** - Observability and monitoring report
   - Has tools: ✅ | Safe outputs: ✅

4. **daily-secrets-analysis** - Security secrets scanning
   - Has tools: ✅ | Safe outputs: ✅

5. **deep-report** - Deep analysis reporting
   - Has tools: ✅ | Safe outputs: ✅

6. **mergefest** - Pull request merge automation
   - Has tools: ✅ | Safe outputs: ✅

7. **pdf-summary** - PDF document summarization
   - Has tools: ✅ | Safe outputs: ✅

8. **repository-quality-improver** - Repository quality analysis and improvements
   - Has tools: ✅ | Safe outputs: ✅

9. **security-guard** - Security monitoring and alerting
   - Has tools: ✅ | Safe outputs: ✅

10. **smoke-claude** - Claude engine smoke test
    - Has tools: ✅ | Safe outputs: ✅

11. **test-workflow** - Test workflow
    - Has tools: ✅ | Safe outputs: ✅

</details>

**Recommended Action:**
Run `make recompile` to regenerate all lock files, or selectively recompile:
```bash
gh aw compile .github/workflows/auto-triage-issues.md
gh aw compile .github/workflows/daily-code-metrics.md
# ... (repeat for all 11 workflows)
```

## Healthy Workflows ✅

**137 workflows (92.6%)** operating normally with up-to-date lock files and no detected issues.

## Systemic Issues

**No systemic issues detected** - The outdated lock files and failing workflow are isolated issues.

## Trends

- **Overall health score**: 78/100 (↓ -19 from 97/100, warning level)
- **New failures this week**: 1 (daily-fact)
- **Fixed issues this week**: 0
- **Average workflow success rate**: 92.6% (137/148 healthy)
- **Workflows needing recompilation**: 11 (7.4%)
- **Compilation success rate**: 100% (148/148)

### Historical Comparison
| Date | Health Score | Critical Issues | Compilation Coverage | Workflow Count | Outdated Locks |
|------|--------------|-----------------|---------------------|----------------|----------------|
| 2026-02-05 | 75/100 | 3 workflows | 100% | - | - |
| 2026-02-06 | 92/100 | 1 workflow | 100% | - | - |
| 2026-02-07 | 94/100 | 1 workflow | 100% | - | - |
| 2026-02-08 | 96/100 | 0 workflows | 100% | 147 | - |
| 2026-02-09 | 97/100 | 0 workflows | 100% | 148 | - |
| 2026-02-10 | 78/100 | 1 workflow | 100% | 148 | 11 |

**Trend**: ⚠️ Significant decline due to outdated locks and 1 failure (daily-fact)

## Ecosystem Statistics

### Engine Distribution
- **Copilot**: 81 workflows (54.7%) - includes "id: copilot" format
- **Claude**: 35 workflows (23.6%) - includes "id: claude" format
- **Codex**: 10 workflows (6.8%) - includes "id: codex" format
- **No engine specified**: 22 workflows (14.9%)

### Feature Adoption
- **Safe outputs enabled**: 140/148 (94.6%)
- **Tools configured**: 138/148 (93.2%)

## Recommendations

### High Priority

1. **Fix daily-fact workflow (P1 - Critical)**
   - Missing `handle_noop_message.cjs` module causing conclusion job failure
   - Create missing file or update workflow configuration
   - Issue already auto-created: [#14763](https://github.com/github/gh-aw/issues/14763)

2. **Recompile outdated workflows (P2 - Medium)**
   - 11 workflows have source changes not reflected in lock files
   - Run `make recompile` to regenerate all lock files
   - Ensures workflows use latest configurations and fixes

### Medium Priority

None identified - ecosystem is otherwise healthy

### Low Priority

None identified

## Actions Taken This Run

- ✅ Created comprehensive health assessment
- ✅ Identified 1 failing workflow with root cause analysis
- ✅ Identified 11 workflows needing recompilation
- ✅ No new issues created (daily-fact already has auto-created issue #14763)
- ✅ Updated shared memory with current health status

## Release Mode Assessment

**Release Mode Status**: ⚠️ WARNING

Given the **release mode** focus on quality and stability:
- ⚠️ **1 workflow failing** (daily-fact - missing JavaScript module)
- ⚠️ **11 workflows outdated** (need recompilation)
- ✅ **100% compilation coverage** maintained
- ✅ **137/148 workflows healthy** (92.6%)
- ✅ **No systemic issues** affecting stability

**Recommendation**: Fix daily-fact module issue and recompile outdated workflows before considering production-ready. Health score drop is significant but fixable.

---
> **Last updated**: 2026-02-10T11:41:59Z  
> **Next check**: Automatic on next trigger or 2026-02-11  
> **Workflow run**: [§21863321000](https://github.com/github/gh-aw/actions/runs/21863321000)
