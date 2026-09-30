# ADR 0001: Use Standard Root-Level Go Layout

## Status

Accepted

## Context

The project needs an executable server under `cmd/server` and reusable skill logic under `pkg/skills`. A `src/` directory was considered as a top-level source directory.

## Decision

Keep Go source at the repository root:

```text
cmd/server/
pkg/skills/
```

Do not introduce `/src`.

## Consequences

- The project follows common Go repository conventions.
- The required package paths remain stable.
- `go build ./cmd/server` works as specified.
- Documentation and tooling do not need special source-root configuration.
