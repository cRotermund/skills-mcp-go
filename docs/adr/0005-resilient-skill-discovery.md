# ADR 0005: Keep Serving Valid Skills When Individual Skills Fail

## Status

Accepted

## Context

A shared skills directory may contain incomplete, malformed, or partially edited skill folders. Failing the entire server startup because of one invalid skill would reduce availability and make incremental authoring difficult.

## Decision

Parse and validate skills independently. Skip malformed skills, emit diagnostics to stderr, and continue serving every valid skill. A later successful rescan may remove a previously valid skill if it has become invalid or disappeared.

Only infrastructure failures affecting the configured root or required watcher setup may prevent startup.

## Consequences

- The server remains useful despite isolated authoring errors.
- Diagnostics and listing behavior must make skipped skills discoverable to operators.
- Repository replacement must distinguish a valid empty snapshot from a failed scan.
