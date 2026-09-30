# ADR 0004: Expose Readable Text Assets Only

## Status

Accepted

## Context

Skills may include helper scripts and tools that are useful as context for an MCP client. Exposing arbitrary files or binary assets would expand the resource surface and require additional content and size semantics.

## Decision

Expose regular, readable text files under `scripts/` and `tools/` as read-only MCP resources. Reject binaries, invalid UTF-8, NUL-containing files, oversized files, traversal, absolute paths, and symlink escapes.

Asset contents are never executed by the server.

## Consequences

- Clients can inspect scripts and source helpers as part of a skill.
- Asset reads have explicit filesystem containment rules.
- Binary tools are out of scope for the first release.
- MIME types can be inferred from known extensions with `text/plain` as a fallback.
