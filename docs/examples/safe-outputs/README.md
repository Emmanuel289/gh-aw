# Safe Output Examples

This directory contains example workflows demonstrating safe output patterns and best practices for GitHub Agentic Workflows.

## Examples

### [Conditional Output Pattern](./conditional-output.md)
Dynamic output type selection based on analysis results, specifically for security vulnerability scanning.

**Pattern:** Route findings by severity
- Critical vulnerabilities → Individual issues
- Medium/low vulnerabilities → Summary discussion
- All findings → Comment on PR

**Use Cases:** Security scanning, code quality analysis, compliance checking

### [Multi-Output Analysis Pattern](./multi-output-analysis.md)
Hierarchical output strategy where a parent discussion contains the overall analysis, with child issues for actionable sub-items.

**Pattern:** Parent summary + child tasks
- Discussion → Comprehensive analysis
- Issues → Individual actionable items
- Comment → Links between items

**Use Cases:** Code quality reports, audit findings, multi-item analysis

### [Fix-or-Report Pattern](./fix-or-report.md)
Progressive approach where the workflow attempts an automated fix first, and falls back to creating an issue if the fix cannot be automated.

**Pattern:** Attempt automation, fallback to manual
- Automated fix possible → Pull request
- Manual intervention needed → Issue
- Always → Summary comment

**Use Cases:** Dependency updates, automated refactoring, configuration fixes

### [Comment Pattern](./comment-pattern.md)
Comment-first approach for status updates with escalation to issues only when necessary.

**Pattern:** Update-focused with escalation
- Always → Add comment with status
- Persistent failure (3+ runs) → Create issue
- Hide older comments → Clean thread

**Use Cases:** CI status reporting, progress updates, monitoring

## Common Patterns

All examples demonstrate:
- ✅ Clear decision logic for output type selection
- ✅ Proper use of `max` limits
- ✅ Cross-referencing between outputs
- ✅ Appropriate use of expiration and cleanup
- ✅ Comprehensive documentation in outputs

## Pattern Selection

| Goal | Pattern | Example |
|------|---------|---------|
| Route by severity/priority | Conditional Output | Security scan results |
| Parent summary + sub-tasks | Multi-Output Analysis | Code quality report |
| Try automated fix first | Fix-or-Report | Dependency updates |
| Status updates + escalation | Comment Pattern | CI status reporting |

## Related Documentation

- **[Safe Outputs Guide](../safe-outputs-guide.md)** - Complete decision tree and best practices
- **[Technical Deep-Dive](../../scratchpad/safe-outputs-patterns.md)** - Implementation details
- **[System Specification](../../scratchpad/safe-outputs-specification.md)** - Formal specification

---

**Last Updated:** 2026-01-31  
**Related Issues:** [#12407](https://github.com/githubnext/gh-aw/issues/12407)
