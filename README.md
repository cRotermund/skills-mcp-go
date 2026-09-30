# Central Skills MCP Server

Central Skills is a Go-native local skills management engine and MCP server. It discovers Markdown-based skills from a configurable directory, validates their YAML frontmatter, tracks semantic versions, and serves skill instructions and text assets over MCP STDIO.

The project is currently in the design and scaffolding phase. The implementation will be added incrementally under `cmd/` and `pkg/`.

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

Go source is kept at the module root rather than under `src/` so the project follows standard Go layout and retains the expected command:

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

The default skill directory is `~/.config/agentic/skills`. Set `CENTRAL_SKILLS_DIR` to use another directory.

## Planned MCP Surface

Resources:

- `skill://{skill-name}` for the complete `SKILL.md`.
- `skill://{skill-name}/scripts/{asset-path}` for readable script assets.
- `skill://{skill-name}/tools/{asset-path}` for readable tool assets.

Prompts:

- `get_skill_prompt`, accepting a skill name, optional version or semver constraint, and optional target workspace context.

Tools:

- `list_available_skills` for version and description metadata.
- `load_skill_context` for retrieving complete skill Markdown through clients that prefer tools.

No server-side execution tool is planned.

## Configuration and Clients

The server will communicate over STDIO. Client-specific configuration examples for Cursor, Claude Desktop, and custom Go or LangGraph hosts will be added when the server implementation is available.

The design and implementation decisions are documented in:

- [`docs/design.md`](docs/design.md)
- [`docs/plan.md`](docs/plan.md)
- [`docs/adr/`](docs/adr/)

## Development Status

The repository currently contains the project scaffold and design documentation. Build, test, and installation commands will be documented here as the implementation lands.
