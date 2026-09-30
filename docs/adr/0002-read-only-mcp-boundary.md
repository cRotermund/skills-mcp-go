# ADR 0002: Keep the MCP Server Read-Only

## Status

Accepted

## Context

An initial proposal included a generic workspace command tool. Direct process execution would be safer than shell parsing, but it would still give an LLM-accessible MCP server significant local execution authority and introduce platform-specific policy concerns.

## Decision

Remove generic command execution from the server. The server will not execute commands, invoke shells, or execute skill scripts.

The MCP surface is limited to skill discovery, context loading, prompts, and read-only resources. A consuming MCP host may provide its own execution tools and approval policy.

## Consequences

- The server has a substantially smaller security boundary.
- Skills can still describe commands and workflows for a capable host.
- Clients that need execution must supply that capability separately.
- The central server remains portable and focused on content delivery.
