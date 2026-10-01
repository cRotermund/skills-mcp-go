# Contributing

Thank you for helping improve Skills MCP Server. This guide explains how to prepare changes for review. For project behavior and user setup, start with the [`README.md`](README.md).

## Before You Start

- Read the relevant section of the README.
- Read [`docs/design.md`](docs/design.md) before changing architecture or MCP behavior.
- Check [`docs/adr/`](docs/adr/) for decisions that already constrain the proposed change.
- Search existing issues and pull requests before opening a duplicate discussion.

## Development Setup

Use the Go version declared by `go.mod`. From the repository root, verify the current implementation with:

```text
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/server
```

The race test may require a working C compiler on platforms where Go's race detector uses CGO.

## Branches

Do not commit or push directly to `master`. Start from an up-to-date `master` branch and create a focused branch using one of these prefixes:

```text
feature/<short-description>
fix/<short-description>
docs/<short-description>
chore/<short-description>
```

Keep each branch limited to one logical change.

## Making Changes

1. Make the smallest change that solves the problem.
2. Add or update tests for behavior changes.
3. Update user-facing documentation when configuration, MCP behavior, or workflows change.
4. Add an ADR when the change creates or supersedes a significant architectural decision.
5. Run the relevant validation commands before opening a pull request.

The server is intentionally read-only. Changes that add command execution, shell invocation, or broader filesystem access require design discussion and an ADR before implementation.

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(optional scope): short imperative description
```

Common types are:

- `feat`: new functionality.
- `fix`: a defect correction.
- `docs`: documentation changes.
- `test`: test-only changes.
- `refactor`: behavior-preserving restructuring.
- `chore`: maintenance and tooling.

Examples:

```text
feat(skills): add semantic version resolution
fix(server): reject missing skill roots
docs: clarify MCP client setup
```

Keep commits cohesive and do not combine unrelated refactors with a feature or fix.

## Pull Requests

Push the feature branch and open a pull request instead of merging directly into `master`. A useful pull request includes:

- A concise summary of what changed and why.
- The user-visible behavior or API impact.
- Tests and validation commands that were run.
- Links to relevant issues or ADRs.
- Known limitations, compatibility concerns, or follow-up work.

Before requesting review, confirm that the branch contains only the intended changes and that the pull request description matches the implementation.

## Documentation

Keep human-facing setup and usage instructions in `README.md`. Keep contribution workflow in this file. Keep architecture, implementation sequencing, and durable design decisions in `docs/`. Agent-specific instructions belong in `AGENTS.md` and should point to these documents rather than copy them.

## Community Expectations

Be specific, constructive, and respectful in issues, reviews, and discussions. Explain the problem before prescribing a solution, and provide enough context for another contributor to reproduce or evaluate the change.
