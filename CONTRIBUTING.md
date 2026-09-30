# Contributing

## Development Principles

- Keep the MCP server read-only. Do not add command execution or shell invocation without a new design review and ADR.
- Prefer small, focused changes with tests at the layer being changed.
- Preserve exact skill Markdown content when serving `SKILL.md` resources.
- Treat malformed skills as diagnostics, not reasons to stop serving valid skills.
- Treat all client-provided resource paths as untrusted input.
- Keep platform-specific behavior explicit and tested.

## Validation

Once implementation is present, changes should pass the repository checks:

```text
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/server
```

Documentation-only changes should still be checked for broken paths, stale commands, and consistency with the design documents.

## Conventional Commits

Commit messages use the Conventional Commits format:

```text
<type>(optional scope): short imperative description
```

Common types include:

- `feat`: new user-visible functionality.
- `fix`: correction of a defect.
- `test`: tests without production behavior changes.
- `docs`: documentation-only changes.
- `refactor`: behavior-preserving code restructuring.
- `chore`: tooling or maintenance work.

Examples:

```text
feat(skills): add semantic version resolution
test(mcp): cover skill resource reads
docs: record read-only server boundary
```

Keep commits cohesive. Do not combine unrelated refactors with feature changes.

## Architecture Decision Records

Create or update an ADR for a significant architectural decision, especially when a change affects:

- The MCP surface or protocol behavior.
- Skill file or URI formats.
- Version resolution semantics.
- Filesystem safety or asset exposure.
- Dependency or project layout strategy.
- The server's authority boundary.

ADRs live in `docs/adr/` and use sequential numeric filenames such as `0006-short-title.md`. An ADR should state the context, decision, consequences, and alternatives considered. Accepted decisions should be updated only when the decision itself changes; use a new ADR when a later decision supersedes an earlier one.

## Pull Requests

Pull requests should describe:

- What changed and why.
- Any user-visible MCP behavior changes.
- Tests and validation performed.
- New or updated ADRs, when applicable.
- Any known compatibility or platform limitations.
