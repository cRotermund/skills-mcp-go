# ADR 0003: Use Dynamic Resources and Semantic Versioning

## Status

Accepted

## Context

Skills can be added or changed while the server is running, and multiple versions of a skill may coexist. Static registration of every concrete resource would require synchronization with the filesystem watcher.

## Decision

Use MCP resource templates for skills and assets. Store multiple semantic versions per skill name and resolve unversioned requests to the latest valid version.

The planned resource forms are:

```text
skill://{skill_name}
skill://{skill_name}/scripts/{+asset_path}
skill://{skill_name}/tools/{+asset_path}
```

Tools and prompts may accept explicit versions or semver constraints.

## Consequences

- Newly discovered skills can be read without re-registering individual resources.
- The repository must be concurrency-safe and version-aware.
- Client support for resource templates must be documented.
- URI-template behavior for nested asset paths must be verified against the selected MCP library.
