# Skills MCP Server

Skills MCP Server is a local, read-only MCP server for discovering Markdown-based skills and making their instructions available to MCP clients over STDIO.

It is useful when you want one skills directory to be discoverable by Cursor, Claude Desktop, or a custom MCP host without giving this server permission to execute commands or scripts.

## Features

- Recursively discovers `SKILL.md` files.
- Validates YAML frontmatter and semantic versions.
- Supports multiple versions of the same skill.
- Resolves unversioned requests to the latest valid version.
- Watches the skills directory for changes.
- Serves complete `SKILL.md` files as Markdown resources.
- Serves readable text assets from `scripts/` and `tools/`.
- Reports malformed skills as diagnostics while continuing to serve valid skills.
- Provides read-only MCP prompts and tools.

The server never executes commands, invokes a shell, or executes skill assets. If a skill describes commands, execution is the responsibility of the consuming MCP host.

## Quick Start: Using the Server

### 1. Create a skills directory

Create a directory for your skills and add one subdirectory per skill:

```text
skills/
└── hello-world/
    └── SKILL.md
```

Add a minimal `SKILL.md`:

```markdown
---
name: hello-world
version: 1.0.0
description: A simple example skill.
---

# Hello World

When this skill is loaded, explain the task clearly and provide a concise answer.
```

### 2. Build the server

From the repository root, build the server binary:

```text
go build -o skills-server ./cmd/server
```

### 3. Start the server

Point `SKILLS_DIR` at the existing skills directory:

```text
SKILLS_DIR=/absolute/path/to/skills skills-server
```

On Windows PowerShell:

```powershell
$env:SKILLS_DIR = "C:\work\skills"
.\skills-server.exe
```

The server starts an MCP STDIO session and makes the `hello-world` skill available through `list_available_skills`, `load_skill_context`, and the `skill://hello-world` resource.

### 4. Connect an MCP client

Configure your MCP client to launch the binary with the same environment variable:

```json
{
  "mcpServers": {
    "skills-mcp": {
      "command": "/absolute/path/to/skills-server",
      "env": {
        "SKILLS_DIR": "/absolute/path/to/skills"
      }
    }
  }
}
```

The server watches the directory after startup. Adding another valid skill directory makes it available after the next filesystem rescan without changing server code.

## Requirements

- Go 1.26 or newer.
- An existing skills directory.

The server does not create the skills directory. It fails to start if the configured directory does not exist or is not a directory.

## Install

Build the server from the repository root:

```text
go build -o skills-server ./cmd/server
```

The resulting `skills-server` binary is the only runtime artifact required by the server.

The Makefile also provides `build`, `test`, `race`, `vet`, `install`, and `dev` targets.

## Skill Format

Each `SKILL.md` starts with YAML frontmatter:

```markdown
---
name: go-concurrency-audit
version: 1.2.0
description: Audit Go code for concurrency hazards.
triggers:
  - goroutine leak
  - concurrency audit
allowed_tools:
  - read_file
---

# Go Concurrency Audit

Markdown instructions follow the frontmatter.
```

The `name`, `version`, and `description` fields are required. `triggers` and `allowed_tools` are optional lists.

Files below `scripts/` and `tools/` are exposed only when they are regular, readable text files. They are never executed by this server.

## MCP Interface

Resources:

- `skill://{skill_name}` returns the complete `SKILL.md`.
- `skill://{skill_name}/scripts/{+asset_path}` returns a readable script asset.
- `skill://{skill_name}/tools/{+asset_path}` returns a readable tool asset.

Prompts:

- `get_skill_prompt` loads a selected skill into conversation context.

Tools:

- `list_available_skills` lists skill names, versions, descriptions, and text assets.
- `load_skill_context` returns the complete `SKILL.md` for a selected skill.

Version arguments accept exact versions and semantic-version constraints. An omitted version selects the latest valid version.

## MCP Client Configuration

An MCP client can launch the server with a configuration equivalent to:

```json
{
  "mcpServers": {
    "skills-mcp": {
      "command": "/absolute/path/to/skills-server",
      "env": {
        "SKILLS_DIR": "/absolute/path/to/skills"
      }
    }
  }
}
```

Custom Go and LangGraph hosts should launch the binary as an MCP STDIO server and use their MCP client adapter to initialize the session and read resources, prompts, and tools.

## Development

Run the repository checks with:

```text
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/server
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) before opening a change. Design details and architectural decisions are in [`docs/`](docs/), including [`docs/design.md`](docs/design.md), [`docs/plan.md`](docs/plan.md), and [`docs/adr/`](docs/adr/).

## Help

For a defect or feature request, open a GitHub issue with a clear description, reproduction steps when applicable, and the relevant Go version and operating system. For implementation questions or proposed changes, see the contribution guide first.

## Maintainers and Contributors

The project is developed collaboratively. Contributions are welcome through reviewed pull requests. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the expected workflow.
