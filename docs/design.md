# Central Skills MCP Runtime Engine Design

## Status

Proposed. This document describes the first implementation scope and the decisions made during planning.

## Goals

The system will provide a local Go service that:

1. Discovers Markdown-based skills from a configurable directory.
2. Parses and validates YAML frontmatter.
3. Maintains an in-memory, semantic-version-aware skill registry.
4. Detects filesystem changes without requiring a server restart.
5. Serves skill instructions and approved text assets through MCP STDIO.
6. Continues operating when individual skills are malformed.

## Non-Goals

The first implementation will not:

- Execute commands or skill scripts.
- Invoke a shell.
- Expose arbitrary filesystem paths.
- Serve binary skill assets.
- Provide an HTTP or network transport.
- Modify skill files.

Execution remains the responsibility of the consuming MCP host, where the host can apply its own user approval and security policies.

## Project Layout

Go source uses the standard root-level layout:

```text
cmd/server/    # executable entrypoint and MCP wiring
pkg/skills/    # reusable skill engine packages
docs/          # design documents and ADRs
testdata/      # checked-in test fixtures
```

The project intentionally does not use `/src`. This preserves standard Go package discovery and the required `go build ./cmd/server` command.

## Configuration

The skill root is configured through `CENTRAL_SKILLS_DIR`.

If unset, the server uses:

```text
~/.config/agentic/skills
```

The application resolves the home directory explicitly because Go filesystem APIs do not expand `~` automatically. The resolved root should be absolute and cleaned before scanning.

## Skill Format

Each skill is a directory containing `SKILL.md`:

```text
skill-directory/
├── SKILL.md
├── scripts/    # optional readable text assets
└── tools/      # optional readable text assets
```

The frontmatter schema is:

```yaml
---
name: string
version: string
description: string
triggers: list[string]
allowed_tools: list[string]
---
```

`name`, `version`, and `description` are required. `triggers` and `allowed_tools` are optional and default to empty lists.

The parser stores both:

- `RawMarkdown`: the exact original file content, including frontmatter.
- `Body`: the Markdown content after the closing frontmatter delimiter.

The complete raw content is used for skill resources and context loading so clients receive the exact authored playbook.

## Discovery and Diagnostics

The scanner recursively searches the configured root for files named `SKILL.md`. Every skill is parsed independently.

Malformed skills are skipped and produce diagnostics. Examples include:

- Missing or unterminated frontmatter.
- Invalid YAML.
- Missing required metadata.
- Invalid semantic versions.
- Duplicate name/version combinations.

Infrastructure failures affecting the configured root or watcher may be fatal during startup. An invalid individual skill is never allowed to prevent valid skills from being served.

Diagnostics are written to stderr through the application's logger. They must never corrupt the MCP JSON-RPC stream on stdout.

## Semantic Versions

The repository supports multiple versions per skill name. Versions are parsed and compared using semantic-version rules rather than lexical string comparison.

The repository will support:

- Exact version lookup.
- Constraint lookup using standard semver expressions such as `>=1.0.0`, `^1.2.0`, and `~1.2.0`.
- Latest-version lookup for unversioned requests.

The latest valid version is the highest satisfying version, excluding pre-releases unless the constraint explicitly permits them according to the selected semver library's rules.

## Repository

The in-memory repository is concurrency-safe. The watcher builds a new valid snapshot and replaces the active snapshot atomically. MCP handlers can read the repository while a rescan is occurring without observing a partially updated registry.

The repository is keyed conceptually as:

```text
skill name -> sorted versions -> skill record
```

Each skill record includes its source directory, raw Markdown, frontmatter, and discovered assets.

## Filesystem Watching

The initial scan loads all valid skills. A filesystem watcher then monitors the configured root and rescans after relevant changes, including creation, modification, rename, and removal events.

Event bursts should be debounced. A full rescan is preferred to complicated incremental state management during the first implementation because it is easier to reason about and preserves consistent diagnostics.

New directories must become watched so newly added nested skills are discovered. A failed rescan should preserve the previous valid snapshot unless the failure is an explicit successful scan that shows a previously valid skill is now invalid or absent.

## MCP Surface

### Skill Resource

```text
skill://{skill-name}
```

This resource returns the complete `SKILL.md` as `text/markdown`. An unversioned request resolves to the latest version. Explicit version selection may be represented through the URI query or the resource resolution contract, subject to the selected MCP library's URI-template behavior.

### Asset Resources

```text
skill://{skill-name}/scripts/{asset-path}
skill://{skill-name}/tools/{asset-path}
```

Asset resources return readable text files only. The server resolves the skill from the repository, then validates the asset path relative to that skill's directory.

The server must reject:

- Absolute paths.
- Parent traversal.
- Encoded traversal after URI decoding.
- Symlink escapes; the first implementation rejects symlinks.
- Files outside `scripts/` and `tools/`.
- Invalid UTF-8 or NUL-containing content.
- Files over the configured maximum asset size.

The server should infer a MIME type from the extension and use `text/plain` as a safe fallback.

### Prompt

`get_skill_prompt` accepts a skill name, optional version or constraint, and optional `target_workspace` context. It injects the exact skill Markdown into a prompt message.

`target_workspace` is contextual text only. The server does not validate, open, or execute against it.

### Tools

`list_available_skills` returns names, versions, descriptions, latest-version information, and asset metadata.

`load_skill_context` returns the complete `SKILL.md` for a selected skill version. It exists for clients that handle tools more naturally than resources.

There is deliberately no generic command execution tool.

## Security Model

The server's authority is limited to reading the configured skills directory and returning validated content. It must not turn client-provided strings into arbitrary filesystem access.

Every asset read is resolved from a repository-owned skill directory and checked for containment before opening. URI decoding and path cleaning must happen before containment validation.

Skill content is untrusted input from the perspective of the server. The server preserves and returns it, but does not execute Markdown instructions or scripts.

## Testing Strategy

Tests cover four layers:

1. Parser, scanner, asset classification, repository, and semver unit tests.
2. Watcher tests using temporary directories and bounded eventual assertions.
3. Direct MCP handler tests against a test repository.
4. A real STDIO subprocess smoke test covering initialization, tools, resources, and asset reads.

Required validation includes:

```text
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/server
```

## Consequences

This design favors a small, auditable read-only service over an all-purpose agent runtime. The tradeoff is that clients must supply their own execution tools when a skill requires commands to be run. That separation avoids platform-specific shell behavior and prevents the central server from becoming an arbitrary process execution boundary.
