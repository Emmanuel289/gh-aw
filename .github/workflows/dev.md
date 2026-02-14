---
on:
  workflow_dispatch:
name: Dev
description: Build and test this project
timeout-minutes: 30
strict: false
sandbox:
  agent: awf
engine: copilot
network:
  allowed:
    - defaults
    - ghcr.io
    - pkg-containers.githubusercontent.com
    - proxy.golang.org
    - sum.golang.org
    - storage.googleapis.com
    - objects.githubusercontent.com
    - codeload.github.com

permissions:
  contents: read
  issues: read
  pull-requests: read

safe-outputs:
  create-pull-request:
    expires: 2h
    title-prefix: "[dev] "
    draft: true
imports:
  - shared/mood.md
---

# Build, Test, and Add Poem

Build and test the gh-aw project, then add a single line poem to poems.txt.

**Requirements:**
1. Run `make build` to build the binary (this handles Go module downloads automatically)
2. Run `make test` to run the test suite
3. Report any failures with details about what went wrong
4. If all steps pass, create a file called poems.txt with a single line poem
5. Create a pull request with the poem

---

## Copilot SDK Engine Example

This workflow can also be used to test the new `copilot-sdk` engine.

The copilot-sdk engine:
- Starts Copilot CLI in headless mode on port 3312
- Uses the Copilot SDK client to communicate with the CLI
- Passes configuration via the `GH_AW_COPILOT_CONFIG` environment variable
- Uses Docker internal host domain for MCP connections

To use the copilot-sdk engine, change the frontmatter:
```yaml
engine: copilot-sdk
```

Optional model configuration:
```yaml
engine:
  id: copilot-sdk
  model: gpt-5.1-pro
```

Optional MCP tools configuration:
```yaml
tools:
  playwright:
    version: v1.41.0
```
