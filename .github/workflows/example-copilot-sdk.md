---
name: Copilot SDK Example
description: Example workflow using the copilot-sdk engine
engine: copilot-sdk

# Trigger workflow manually or on push
on:
  workflow_dispatch:
  push:
    branches:
      - main

# Optional: Specify a custom model
# engine:
#   model: gpt-5.1-pro

# Optional: Configure MCP tools
# tools:
#   playwright:
#     version: v1.41.0
---

# Copilot SDK Example

This is an example workflow using the new `copilot-sdk` engine.

The copilot-sdk engine:
- Starts Copilot CLI in headless mode on port 3312
- Uses the Copilot SDK client to communicate with the CLI
- Passes configuration via the `GH_AW_COPILOT_CONFIG` environment variable
- Uses Docker internal host domain for MCP connections

## Task

Please analyze this repository and provide a summary of:
1. The main programming language used
2. The purpose of the project
3. Any notable files or directories

Keep the response brief and focused.
