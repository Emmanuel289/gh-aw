# Playwright MCP Startup Investigation

**Date**: 2026-02-03  
**Issue**: Playwright MCP "calling initialize: EOF" error during gateway startup  
**Workflow**: daily-multi-device-docs-tester  
**Run ID**: 21638629201  
**Status**: ✅ Resolved - Cosmetic error, no functional impact

## Executive Summary

The Playwright MCP server reports an "EOF" error during MCP Gateway startup, but this is a **cosmetic issue** that does not affect workflow functionality. Playwright successfully starts and operates correctly when the agent actually connects.

## Timeline

| Time | Event | Status |
|------|-------|--------|
| 16:35:03 | Gateway pre-registration attempts to initialize Playwright | ❌ Failed with EOF |
| 16:35:08 | Gateway health check | ✅ Reports Playwright running |
| 16:35:48 | Claude agent connects to Playwright via HTTP | ✅ Connection successful |
| 16:35:48-16:39:39 | Playwright tools used for multi-device testing | ✅ All operations succeed |
| 16:39:39 | Workflow completes with test results | ✅ Success |

## Root Cause

### The Problem

During startup, the MCP Gateway attempts to **eagerly initialize** all configured MCP servers to pre-register their tools. This process:

1. Launches the Playwright Docker container
2. Attempts immediate stdio handshake via MCP protocol
3. Fails with "calling 'initialize': EOF" 
4. Reports the error but continues gateway startup

### Why It Fails

The Playwright MCP container needs a brief moment to:
- Initialize the Node.js process
- Set up browser binaries
- Start the MCP protocol handler

**Gateway Timeout**: The initialization happens too quickly, causing an EOF when the container hasn't fully started its stdio handler.

### Why It Works Anyway

The MCP Gateway uses **lazy initialization** as a fallback:

1. Even though pre-registration fails, the gateway **keeps the `/mcp/playwright` route active**
2. When the agent connects later (at 16:35:48, 45 seconds after startup)
3. Playwright is fully initialized and responds correctly
4. All tools are available and functional

## Evidence

### 1. Error During Startup

```log
2026/02/03 16:35:03 [LAUNCHER] Starting MCP server: playwright
2026/02/03 16:35:03 ❌ MCP Connection Failed:
2026/02/03 16:35:03    Error: calling "initialize": EOF
2026/02/03 16:35:03    ⚠️  Process started but terminated unexpectedly
```

### 2. Gateway Continues Successfully

```log
2026/02/03 16:35:03 Starting MCPG in ROUTED mode on 0.0.0.0:80
2026/02/03 16:35:03 Routes: /mcp/<server> where <server> is one of: [safeoutputs github playwright]
2026/02/03 16:35:03 Registered route: /mcp/playwright
```

### 3. Agent Connection Succeeds

```log
2026-02-03T16:35:48Z [DEBUG] MCP server "playwright": Successfully connected in 85ms
2026-02-03T16:35:48Z [DEBUG] MCP server "playwright": Connection established
```

### 4. Tools Actually Work

The workflow successfully:
- Tested 10+ device configurations (mobile, tablet, desktop)
- Navigated to multiple pages
- Took screenshots
- Ran accessibility checks
- Generated comprehensive test results

## Impact Assessment

### User Impact

**None** - The error message appears in logs but:
- ✅ Workflows execute successfully
- ✅ Playwright tools are fully functional
- ✅ Results are generated correctly
- ✅ No retry or manual intervention needed

### Operational Impact

**Low** - May cause confusion:
- ⚠️ Error messages in gateway logs look alarming
- ⚠️ Pre-registration failure might mask real issues
- ⚠️ Monitoring alerts could trigger incorrectly

## Recommendations

### Short Term (Completed)

1. ✅ **Document this behavior** - This investigation serves as documentation
2. ✅ **Verify no functional impact** - Confirmed through artifact analysis

### Medium Term (Proposed)

1. **Improve error messaging** in MCP Gateway:
   ```
   ⚠️ Pre-registration failed for 'playwright' (will retry on first connection)
   ```
   Instead of:
   ```
   ❌ FAILED to launch server 'playwright'
   ```

2. **Add retry logic** with exponential backoff

3. **Distinguish error types**:
   - "Pre-registration timeout" (non-critical)
   - "Server unavailable" (critical)

## Conclusion

The "Playwright MCP not properly started" error is a **false alarm**. The MCP Gateway's eager initialization strategy encounters a timing issue with Playwright's container startup, but the lazy initialization fallback ensures everything works correctly when the agent actually connects.

**No code changes are required** for functionality. Improvements to error messaging and retry logic would enhance observability but are not critical.
