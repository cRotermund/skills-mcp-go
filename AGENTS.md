# Agent Guidance

This file is an agent-facing index and operating policy. It intentionally does not duplicate human setup, contribution workflow, or architecture prose.

## Read First

- Project purpose, setup, usage, MCP interface, and support: [`README.md`](README.md).
- Branches, testing, commits, pull requests, and documentation ownership: [`CONTRIBUTING.md`](CONTRIBUTING.md).
- Architecture and security model: [`docs/design.md`](docs/design.md).
- Implementation sequence and definition of done: [`docs/plan.md`](docs/plan.md).
- Durable architectural decisions: [`docs/adr/`](docs/adr/).

Read the relevant source document before making changes. Treat those documents as the source of truth rather than copying their contents into this file.

## Repository Map

- `cmd/server/` contains the executable entrypoint and MCP wiring.
- `pkg/skills/` contains reusable skill parsing, scanning, repository, and watcher logic.
- `testdata/skills/` contains stable skill fixtures.

## Agent Operating Rules

- Do not change architecture without first reading `docs/design.md` and `docs/plan.md`.
- Do not silently broaden the server's authority or MCP surface.
- Preserve user changes in the worktree; never revert unrelated work.
- Use focused edits and add tests for changed behavior.
- Keep protocol output and diagnostic output separated according to the implementation's transport contract.
- Do not stage, commit, push, or open a pull request unless the user explicitly requests that action.
- Before reporting implementation complete, run the applicable validation commands from `CONTRIBUTING.md` and disclose any environment limitations.

## Documentation Ownership

When documentation changes are needed:

- Update `README.md` for human users.
- Update `CONTRIBUTING.md` for human contributors.
- Update `docs/` for architecture, plans, and decisions.
- Update this file only for agent-specific navigation or operating rules.
