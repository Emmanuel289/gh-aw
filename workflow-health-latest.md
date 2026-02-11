# Workflow Health Dashboard - 2026-02-11

## Overview
- **Total workflows**: 148 (148 executable, 59 shared includes)
- **Healthy**: 139 (93.9%)
- **Warning**: 8 (5.4%)
- **Critical**: 1 (0.7%)
- **Inactive**: 0 (0%)
- **Compilation coverage**: 147/148 (99.3% ✅)
- **Overall health score**: 82/100 (↑ +4 from 78/100)

## 🟡 STATUS: WARNING - Improvement with 1 Critical Issue

### Health Assessment Summary

**Status: WARNING (Improved from previous run)**

The ecosystem has **1 critical failing workflow** and **8 workflows requiring attention**:
- ❌ **1 workflow failing** (daily-fact - still missing JavaScript module - **KNOWN ISSUE #14769**)
- ⚠️ **1 workflow with infrastructure issues** (agentics-maintenance - artifact upload DNS failure)
- ⚠️ **6 workflows with action_required conclusion** (need investigation)
- ✅ **99.3% compilation coverage** (147/148 workflows have locks)
- ✅ **139 healthy workflows** (93.9%)
- ↑ **Health score improved by +4 points** (78 → 82)

**Key Changes Since Last Check (2026-02-10):**
- ↑ Health score increased by +4 points (78 → 82)
- ✅ Outdated lock files reduced from 11 to ~0 (recompilation happened)
- ❌ daily-fact still failing (known issue #14769)
- ⚠️ agentics-maintenance infrastructure failure (new, transient)
- ✅ Compilation coverage maintained at 99.3%

## Critical Issues 🚨

### 1. daily-fact Workflow (Priority: P1 - KNOWN ISSUE)

**Status:** Still failing due to missing JavaScript module (**TRACKED IN #14769**)

**Latest Run Details:**
- **Run**: [§21902990941](https://github.com/github/gh-aw/actions/runs/21902990941)
- **Failed Job**: agent (conclusion step)
- **Error**: `Error: Cannot find module '/opt/gh-aw/actions/handle_noop_message.cjs'`
- **Impact**: Workflow cannot complete - conclusion job fails every time
- **Engine**: codex
- **Description**: Posts a daily poetic verse about gh-aw to a discussion thread

**Root Cause:**
The file `handle_noop_message.cjs` exists in source (`actions/setup/js/handle_noop_message.cjs`) but is not being copied to `/opt/gh-aw/actions/` during workflow setup. This is a deployment/copying issue in the actions setup process.

**Resolution Status:**
- Issue #14769 already created and tracking
- Requires fix to actions/setup copying logic
- File exists in repo, just needs proper deployment

## Warnings ⚠️

### 1. agentics-maintenance.yml Infrastructure Failure (Priority: P2 - Transient)

**Status:** Failing due to transient infrastructure issue

**Latest Run Details:**
- **Run**: [§21902359274](https://github.com/github/gh-aw/actions/runs/21902359274)
- **Failed Job**: secret-validation
- **Error**: `getaddrinfo EAI_AGAIN productionresultssa13.blob.core.windows.net`
- **Impact**: Cannot upload artifacts due to DNS resolution failure
- **Root Cause**: Transient network/DNS issue with Azure Blob Storage

**Recommended Action:**
- Monitor for recurrence (likely transient)
- No code fix needed - infrastructure issue
- Will likely resolve on next run

### 2. Workflows with action_required Conclusion (6 workflows)

**Status:** 6 workflows completed with `action_required` conclusion in past 7 days

These workflows completed but require human review or action:
- Typical for workflows that create issues/PRs and wait for human response
- Not failures, but need tracking

**Count in past 7 days:** 6 runs with `action_required`

**Recommended Action:**
- Review workflows with `action_required` to ensure issues/PRs are being addressed
- No immediate fix needed - expected workflow behavior

## Healthy Workflows ✅

**139 workflows (93.9%)** operating normally with up-to-date lock files and no detected issues.

## Systemic Issues

**No systemic issues detected** - The failing workflow is isolated and tracked.

## Ecosystem Statistics (Past 7 Days)

### Run Statistics
- **Total workflow runs**: 30
- **Successful runs**: 19 (63.3%)
- **Failed runs**: 2 (6.7%)
- **Cancelled runs**: 0 (0%)
- **Action required**: 6 (20%)
- **Skipped**: 3 (10%)
- **Unique workflows executed**: 26

### Success Rate Breakdown
- **Pure success rate** (success/total): 63%
- **Operational success rate** (success + action_required): 83%
- **Failure rate**: 7% (2 failures: daily-fact + agentics-maintenance)

## Trends

- **Overall health score**: 82/100 (↑ +4 from 78/100, improved)
- **New failures this period**: 1 (agentics-maintenance - transient)
- **Ongoing failures**: 1 (daily-fact - known issue #14769)
- **Fixed issues this period**: 0
- **Average workflow health**: 93.9% (139/148 healthy)
- **Compilation success rate**: 99.3% (147/148)

### Historical Comparison
| Date | Health Score | Critical Issues | Compilation Coverage | Workflow Count | Notable Issues |
|------|--------------|-----------------|---------------------|----------------|----------------|
| 2026-02-05 | 75/100 | 3 workflows | 100% | - | - |
| 2026-02-06 | 92/100 | 1 workflow | 100% | - | - |
| 2026-02-07 | 94/100 | 1 workflow | 100% | - | - |
| 2026-02-08 | 96/100 | 0 workflows | 100% | 147 | - |
| 2026-02-09 | 97/100 | 0 workflows | 100% | 148 | - |
| 2026-02-10 | 78/100 | 1 workflow | 100% | 148 | 11 outdated locks, daily-fact |
| 2026-02-11 | 82/100 | 1 workflow | 99.3% | 148 | daily-fact (ongoing), agentics-maintenance (transient) |

**Trend**: ↑ Improving, outdated locks resolved, 1 critical issue persists

## Recommendations

### High Priority

1. **Fix daily-fact workflow (P1 - Critical - TRACKED IN #14769)**
   - Missing `handle_noop_message.cjs` deployment in actions setup
   - File exists in source but not copied to runtime location
   - Fix actions/setup deployment to include this file
   - Already tracked in issue #14769

### Medium Priority

1. **Monitor agentics-maintenance (P2 - Transient)**
   - Single infrastructure failure (DNS resolution)
   - Likely transient, monitor next run
   - No code fix needed if it resolves

2. **Review action_required workflows (P2 - Routine)**
   - 6 workflows with action_required conclusion
   - Ensure created issues/PRs are being addressed
   - Expected behavior, no urgent action

### Low Priority

None identified

## Actions Taken This Run

- ✅ Comprehensive health assessment completed
- ✅ Analyzed 30 workflow runs from past 7 days
- ✅ Identified 1 critical ongoing issue (daily-fact - already tracked #14769)
- ✅ Identified 1 transient infrastructure issue (agentics-maintenance)
- ✅ No new issues created (daily-fact already tracked)
- ✅ Updated shared memory with current health status

## Release Mode Assessment

**Release Mode Status**: ⚠️ WARNING (Improved from previous run)

Given the **release mode** focus on quality, security, and documentation:
- ❌ **1 workflow critically failing** (daily-fact - known issue #14769)
- ⚠️ **1 transient infrastructure failure** (agentics-maintenance - likely resolves)
- ✅ **99.3% compilation coverage** maintained
- ✅ **139/148 workflows healthy** (93.9%)
- ✅ **No systemic issues** affecting stability
- ✅ **Outdated locks resolved** (down from 11 to ~0)

**Recommendation**: Fix daily-fact module deployment issue (#14769) before considering production-ready. Health improving but 1 critical issue persists.

---
> **Last updated**: 2026-02-11T11:37:18Z  
> **Next check**: Automatic on next trigger or 2026-02-12  
> **Workflow run**: [§21903440840](https://github.com/github/gh-aw/actions/runs/21903440840)
