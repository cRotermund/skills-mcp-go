# Repository Guidance

## Structure

- `cmd/server/` contains the thin executable entrypoint and MCP server wiring.
- `pkg/skills/` contains reusable parsing, scanning, repository, and watcher logic.
- `docs/` contains the design, implementation plan, and ADRs.
- `testdata/skills/` contains stable representative skill fixtures.

Do not move Go code under `src/`; the root-level layout preserves standard Go conventions and the required `go build ./cmd/server` command.

## Architectural Boundaries

- The server is read-only.
- It does not execute commands or scripts.
- It does not invoke a shell.
- It serves `SKILL.md` and approved readable text assets only.
- It must not expose arbitrary filesystem paths.
- Invalid individual skills produce diagnostics and are skipped; they do not prevent valid skills from loading.
- MCP transport output must remain on stdout. Diagnostic logging belongs on stderr.

## Skill Rules

- Parse YAML frontmatter with `gopkg.in/yaml.v3`.
- Preserve the complete original `SKILL.md` text.
- Use semantic versioning for skill versions and constraints.
- Support multiple versions per skill name.
- Resolve an unversioned skill request to the latest valid version.
- Expose only regular, readable text files below `scripts/` and `tools/`.
- Reject traversal, absolute paths, symlink escapes, binary content, invalid UTF-8, and oversized assets.

## Change Process

- Use Conventional Commits.
- Add an ADR for significant architectural decisions.
- Keep changes focused and add tests for changed behavior.
- Before considering implementation complete, run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./cmd/server`.

Read `docs/design.md` and `docs/plan.md` before changing the architecture.
