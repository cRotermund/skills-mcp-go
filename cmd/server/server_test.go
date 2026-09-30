package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cRotermund/skills-mcp-go/pkg/skills"
	"github.com/mark3labs/mcp-go/server"
)

func TestNewMCPServerRegistersReadOnlySurface(t *testing.T) {
	repository, err := skills.NewRepository([]skills.Skill{{
		Frontmatter: skills.Frontmatter{Name: "demo", Version: "1.0.0", Description: "Demo"},
		RawMarkdown: "---\nname: demo\nversion: 1.0.0\ndescription: Demo\n---\n",
	}})
	if err != nil {
		t.Fatal(err)
	}
	mcpServer := NewMCPServer(repository)
	tools := mcpServer.ListTools()
	if len(tools) != 2 {
		t.Fatalf("registered %d tools, want 2", len(tools))
	}
	if tools["execute_workspace_command"] != nil {
		t.Fatal("execution tool must not be registered")
	}
	if mcpServer.ListPrompts()["get_skill_prompt"] == nil {
		t.Fatal("get_skill_prompt was not registered")
	}
}

func TestResourceURIParsing(t *testing.T) {
	name, constraint, err := skillURI("skill://demo?version=%5E1.2.0")
	if err != nil || name != "demo" || constraint != "^1.2.0" {
		t.Fatalf("skillURI = %q, %q, %v", name, constraint, err)
	}
	name, constraint, kind, path, err := assetURI("skill://demo/scripts/helpers/check.sh")
	if err != nil || name != "demo" || constraint != "" || kind != "scripts" || path != "scripts/helpers/check.sh" {
		t.Fatalf("assetURI = %q, %q, %q, %q, %v", name, constraint, kind, path, err)
	}
	if _, _, _, _, err := assetURI("skill://demo/scripts/../SKILL.md"); err == nil {
		t.Fatal("assetURI accepted traversal")
	}
}

func TestMCPServerHandlesStdioRequests(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "demo")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: demo\nversion: 1.0.0\ndescription: Demo\n---\n# Demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "scripts", "check.sh"), []byte("echo asset\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scanner := skills.NewScanner()
	scan, err := scanner.Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := skills.NewRepository(scan.Skills)
	if err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"resources/templates/list","params":{}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"load_skill_context","arguments":{"skill_name":"demo"}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"resources/read","params":{"uri":"skill://demo/scripts/check.sh"}}`,
	}, "\n") + "\n"

	var output bytes.Buffer
	stdio := server.NewStdioServer(NewMCPServer(repository))
	if err := stdio.Listen(context.Background(), strings.NewReader(input), &output); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	responses := make(map[string]json.RawMessage)
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		var response struct {
			ID     json.RawMessage `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("invalid JSON response %q: %v", line, err)
		}
		responses[string(response.ID)] = response.Result
	}
	for _, id := range []string{"1", "2", "3", "4", "5"} {
		if len(responses[id]) == 0 {
			t.Fatalf("missing response for request %s; output=%s", id, output.String())
		}
	}
	var initialize map[string]any
	if err := json.Unmarshal(responses["1"], &initialize); err != nil {
		t.Fatal(err)
	}
	if initialize["serverInfo"] == nil || initialize["capabilities"] == nil {
		t.Fatalf("initialize result = %#v", initialize)
	}
	if !strings.Contains(output.String(), "# Demo") {
		t.Fatalf("load_skill_context response did not contain skill content: %s", output.String())
	}
	if !strings.Contains(output.String(), "echo asset") {
		t.Fatalf("asset resource response did not contain asset content: %s", output.String())
	}
}

func TestBuiltServerHandlesInitialization(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: demo\nversion: 1.0.0\ndescription: Demo\n---\n# Demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(t.TempDir(), "skills-server.exe")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, output)
	}

	command := exec.Command(binaryPath)
	command.Env = append(os.Environ(), "SKILLS_DIR="+root)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke-test","version":"1"}}}` + "\n"
	if _, err := stdin.Write([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if err := stdin.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("server exited with %v; output=%s", err, output)
	}
	if !strings.Contains(string(output), `"serverInfo"`) {
		t.Fatalf("initialize response missing serverInfo: %s", output)
	}
}
