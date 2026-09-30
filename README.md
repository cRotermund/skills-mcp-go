# Skills MCP Server

Skills MCP Server is a Go-native local skills management engine and MCP server. It discovers Markdown-based skills from a configurable directory, validates their YAML frontmatter, tracks semantic versions, and serves skill instructions and text assets over MCP STDIO.

The initial implementation is being developed under `cmd/` and `pkg/`. The server is designed as a local, read-only process with no command execution authority.

## Scope

The first release is intentionally read-only:

- Discover `SKILL.md` files recursively.
- Preserve and serve the complete Markdown skill document.
- Track multiple semantic versions of a skill.
- Expose readable text files under `scripts/` and `tools/` as resources.
- Provide MCP resources, prompts, and read-only discovery tools.
- Continue serving valid skills when individual skills are malformed.

The server will not execute commands, invoke shells, or execute skill scripts. A consuming MCP host may provide its own execution capabilities when appropriate.

## Planned Layout

```text
.
├── cmd/server/       # MCP server entrypoint
├── pkg/skills/       # Parser, scanner, repository, and watcher
├── docs/              # Design, implementation plan, and ADRs
└── testdata/skills/  # Representative skill fixtures
```

Build the server with:

```text
go build ./cmd/server
```

## Skill Format

Each skill is stored in its own directory:

```text
go-concurrency-audit/
├── SKILL.md
├── scripts/          # Optional readable text assets
└── tools/            # Optional readable text assets
```

`SKILL.md` begins with YAML frontmatter:

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

The default skill directory is `~/.config/agentic/skills`. Set `SKILLS_DIR` to use another directory.

## Planned MCP Surface

Resources:

- `skill://{skill_name}` for the complete `SKILL.md`.
- `skill://{skill_name}/scripts/{+asset_path}` for readable script assets.
- `skill://{skill_name}/tools/{+asset_path}` for readable tool assets.

Prompts:

- `get_skill_prompt`, accepting a skill name, optional version or semver constraint, and optional target workspace context.

Tools:

- `list_available_skills` for version and description metadata.
- `load_skill_context` for retrieving complete skill Markdown through clients that prefer tools.

No server-side execution tool is planned.

## Configuration and Clients

The server communicates over STDIO. Set the skill directory before starting it:

```text
SKILLS_DIR=/absolute/path/to/skills skills-server
```

On Windows PowerShell:

```powershell
$env:SKILLS_DIR = "C:\work\skills"
.\skills-server.exe
```

The default directory is `~/.config/agentic/skills`.

The server can be registered with an MCP client using a configuration equivalent to:

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

Custom Go and LangGraph hosts should launch the binary as an MCP STDIO server and use their host's MCP client adapter to initialize the session, list tools/resources, and read skill context. The server does not execute commands described by a skill; execution, if supported, belongs to the consuming host.

## Development Commands

```text
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/server
```

The Makefile provides equivalent `build`, `test`, `race`, `vet`, `install`, and `dev` targets.

The design and implementation decisions are documented in:

- [`docs/design.md`](docs/design.md)
- [`docs/plan.md`](docs/plan.md)
- [`docs/adr/`](docs/adr/)

## Development Status

The initial read-only server implementation, tests, build tooling, and design documentation are present. Additional client compatibility and operational documentation can be expanded as implementation experience accumulates.
