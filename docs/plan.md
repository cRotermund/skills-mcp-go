# Central Skills MCP Runtime Engine Implementation Plan

## Scope

This plan implements the read-only Go MCP server described in `docs/design.md`. It does not include command execution, shell invocation, or binary asset serving.

## Phase 1: Module and Project Setup

1. Create `go.mod` using the selected current Go toolchain.
2. Pin `github.com/mark3labs/mcp-go` to a compatible release.
3. Add `gopkg.in/yaml.v3` for frontmatter parsing.
4. Add a semantic-version library.
5. Add a filesystem watcher dependency.
6. Add the Makefile targets for build, test, install, and development after the executable exists.

The dependency versions must be checked together because current `mcp-go` releases may require a newer Go version than the original 1.22 baseline.

## Phase 2: Models and Parser

Create the skill model and parser under `pkg/skills/`.

Implement:

- Frontmatter structs.
- Skill and asset structs.
- Parsing from bytes for unit-test isolation.
- Parsing from a path for production use.
- Required-field validation.
- Exact `RawMarkdown` preservation.
- Markdown body extraction.

Add unit tests for valid documents, malformed delimiters, invalid YAML, missing fields, line endings, and raw-content preservation.

## Phase 3: Scanner and Asset Discovery

Implement recursive discovery of `SKILL.md` files.

For every parsed skill:

1. Walk only `scripts/` and `tools/`.
2. Consider regular files only.
3. Reject symlinks for the first implementation.
4. Enforce the asset-size limit.
5. Reject invalid UTF-8 and NUL-containing files.
6. Record normalized relative paths and MIME types.

Return valid skills and diagnostics separately so one malformed skill does not stop the scan.

Add temporary-directory tests covering nested skills, malformed files, assets, binary detection, path normalization, and duplicate names.

## Phase 4: Repository and Semantic Versioning

Implement a concurrency-safe in-memory repository.

Required operations:

- Replace the active snapshot.
- List all valid skills and versions.
- Get the latest version for a name.
- Get an exact version.
- Resolve a semver constraint.
- Return asset metadata for a selected version.

Sort versions semantically and reject duplicate name/version pairs. Add repository tests, including concurrent reads during replacement and pre-release behavior.

## Phase 5: Filesystem Watcher

Implement watcher startup against the configured root.

The watcher should:

- Monitor nested directories.
- Add watches for new directories.
- React to create, write, rename, and remove events.
- Debounce event bursts.
- Rescan into a new snapshot.
- Replace the repository only with the scan result.
- Log diagnostics without stopping the server.

Use bounded eventual assertions in watcher tests rather than relying on exact event counts.

## Phase 6: MCP Server Wiring

Keep `cmd/server/main.go` thin. Put server construction in a testable function if needed.

Startup sequence:

1. Read and resolve `CENTRAL_SKILLS_DIR`.
2. Perform the initial scan.
3. Log diagnostics.
4. Create the repository.
5. Start the watcher.
6. Construct `mcp.NewMCPServer`.
7. Register the skill resource template.
8. Register the script and tool asset resource templates.
9. Register `get_skill_prompt`.
10. Register `list_available_skills`.
11. Register `load_skill_context`.
12. Start `server.ServeStdio`.

All diagnostics must go to stderr. Stdout is reserved for MCP protocol traffic.

## Phase 7: Resource and Prompt Behavior

Implement handlers that resolve the current repository at request time.

Test:

- Latest-version resolution.
- Explicit version or constraint resolution.
- Exact raw Markdown output.
- Prompt construction with and without `target_workspace`.
- Unknown skill and version errors.
- Asset MIME types.
- Nested asset paths.
- Traversal and containment rejection.

Confirm the selected `mcp-go` release's URI-template behavior for nested asset paths before finalizing the public URI syntax. Use its supported path expansion or encoded representation rather than custom protocol behavior.

## Phase 8: Integration and Protocol Tests

Add a real STDIO subprocess test that:

1. Starts the server with a temporary skill root.
2. Completes MCP initialization.
3. Lists tools and resource templates.
4. Calls `list_available_skills`.
5. Reads a skill resource.
6. Reads a text asset resource.
7. Confirms malformed skills are absent.

Add a dynamic discovery test that creates, modifies, and removes skills while the server is running.

## Phase 9: Documentation and Packaging

Complete:

- Cursor configuration example.
- Claude Desktop configuration example.
- Go and LangGraph adapter guidance.
- Skill authoring guide.
- Makefile build, test, install, and development targets.
- Static binary output documentation.

The README must clearly state that the server does not execute commands or scripts.

## Definition of Done

The implementation is ready for its first commit when:

- `go build ./cmd/server` succeeds.
- `go test ./...` succeeds.
- `go test -race ./...` succeeds.
- `go vet ./...` succeeds.
- A real STDIO initialize request succeeds.
- Valid skills are dynamically discoverable.
- Malformed skills produce diagnostics without stopping service.
- Skill Markdown and readable assets are available through MCP.
- Asset reads cannot escape their owning skill directory.
- No server code executes commands or scripts.
- Documentation and ADRs match the implemented behavior.
