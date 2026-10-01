package main

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cRotermund/skills-mcp-go/pkg/skills"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "skills-server"
	serverVersion = "0.1.0"
)

func NewMCPServer(repository *skills.Repository) *server.MCPServer {
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithRecovery(),
		server.WithResourceRecovery(),
		server.WithInputSchemaValidation(),
		server.WithStrictInputSchemaDefault(),
		server.WithInstructions("Read-only access to discovered skills and their text assets."),
	)

	mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"skill://{skill_name}",
			"Skill instructions",
			mcp.WithTemplateDescription("The complete SKILL.md document."),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		skillResourceHandler(repository),
	)
	mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"skill://{skill_name}/scripts/{+asset_path}",
			"Skill script asset",
			mcp.WithTemplateDescription("A readable text asset below scripts/."),
			mcp.WithTemplateMIMEType("text/plain"),
		),
		assetResourceHandler(repository),
	)
	mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"skill://{skill_name}/tools/{+asset_path}",
			"Skill tool asset",
			mcp.WithTemplateDescription("A readable text asset below tools/."),
			mcp.WithTemplateMIMEType("text/plain"),
		),
		assetResourceHandler(repository),
	)

	mcpServer.AddPrompt(skillPrompt(), skillPromptHandler(repository))
	mcpServer.AddTool(listSkillsTool(), listSkillsHandler(repository))
	mcpServer.AddTool(loadSkillTool(), loadSkillHandler(repository))
	return mcpServer
}

func skillResourceHandler(repository *skills.Repository) server.ResourceTemplateHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		name, constraint, err := skillURI(request.Params.URI)
		if err != nil {
			return nil, err
		}
		skill, err := repository.Resolve(name, constraint)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/markdown",
			Text:     skill.RawMarkdown,
		}}, nil
	}
}

func assetResourceHandler(repository *skills.Repository) server.ResourceTemplateHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		name, constraint, kind, relativePath, err := assetURI(request.Params.URI)
		if err != nil {
			return nil, err
		}
		if kind != skills.AssetScripts && kind != skills.AssetTools {
			return nil, fmt.Errorf("unsupported asset directory %q", kind)
		}
		skill, err := repository.Resolve(name, constraint)
		if err != nil {
			return nil, err
		}
		asset, text, err := skills.ReadAssetText(skill, relativePath)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: asset.MIMEType,
			Text:     text,
		}}, nil
	}
}

func skillPrompt() mcp.Prompt {
	return mcp.NewPrompt(
		"get_skill_prompt",
		mcp.WithPromptDescription("Load a selected skill into conversation context."),
		mcp.WithArgument("skill_name", mcp.ArgumentDescription("Skill name"), mcp.RequiredArgument()),
		mcp.WithArgument("version", mcp.ArgumentDescription("Optional exact version or semver constraint")),
		mcp.WithArgument("target_workspace", mcp.ArgumentDescription("Optional contextual workspace path")),
	)
}

func skillPromptHandler(repository *skills.Repository) server.PromptHandlerFunc {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := request.Params.Arguments["skill_name"]
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("skill_name is required")
		}
		constraint := request.Params.Arguments["version"]
		skill, err := repository.Resolve(name, constraint)
		if err != nil {
			return nil, err
		}
		content := "Use the following skill instructions exactly as the skill context:\n\n" +
			"--- BEGIN SKILL ---\n" + skill.RawMarkdown +
			"\n--- END SKILL ---"
		if workspace := request.Params.Arguments["target_workspace"]; workspace != "" {
			content += "\n\nTarget workspace context: " + workspace
		}
		return mcp.NewGetPromptResult(
			fmt.Sprintf("Skill %s@%s", skill.Name(), skill.Version()),
			[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(content))},
		), nil
	}
}

type skillSummary struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	IsLatest    bool           `json:"is_latest"`
	Assets      []skills.Asset `json:"assets"`
}

func listSkillsTool() mcp.Tool {
	return mcp.NewTool(
		"list_available_skills",
		mcp.WithDescription("List discovered skills, versions, descriptions, and readable assets."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
	)
}

func listSkillsHandler(repository *skills.Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		all := repository.List()
		latest := make(map[string]string)
		for _, skill := range all {
			if _, exists := latest[skill.Name()]; !exists {
				latest[skill.Name()] = skill.Version()
			}
		}
		summaries := make([]skillSummary, 0, len(all))
		for _, skill := range all {
			summaries = append(summaries, skillSummary{
				Name:        skill.Name(),
				Version:     skill.Version(),
				Description: skill.Frontmatter.Description,
				IsLatest:    latest[skill.Name()] == skill.Version(),
				Assets:      skill.Assets,
			})
		}
		return mcp.NewToolResultJSON(summaries)
	}
}

func loadSkillTool() mcp.Tool {
	return mcp.NewTool(
		"load_skill_context",
		mcp.WithDescription("Return the complete SKILL.md for a selected skill version."),
		mcp.WithString("skill_name", mcp.Required(), mcp.Description("Skill name")),
		mcp.WithString("version", mcp.Description("Optional exact version or semver constraint")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
	)
}

func loadSkillHandler(repository *skills.Repository) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("skill_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		skill, err := repository.Resolve(name, request.GetString("version", ""))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(skill.RawMarkdown), nil
	}
}

func skillURI(raw string) (string, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("parse skill URI: %w", err)
	}
	if parsed.Scheme != "skill" || parsed.Host == "" || parsed.Path != "" {
		return "", "", fmt.Errorf("invalid skill URI %q", raw)
	}
	constraint, err := queryValue(parsed, "version")
	if err != nil {
		return "", "", err
	}
	return parsed.Host, constraint, nil
}

func assetURI(raw string) (string, string, string, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", "", "", fmt.Errorf("parse asset URI: %w", err)
	}
	if parsed.Scheme != "skill" || parsed.Host == "" {
		return "", "", "", "", fmt.Errorf("invalid asset URI %q", raw)
	}
	path := parsed.EscapedPath()
	decoded, err := url.PathUnescape(path)
	if err != nil {
		return "", "", "", "", fmt.Errorf("decode asset URI: %w", err)
	}
	parts := strings.SplitN(strings.TrimPrefix(decoded, "/"), "/", 2)
	if len(parts) != 2 || (parts[0] != skills.AssetScripts && parts[0] != skills.AssetTools) || parts[1] == "" {
		return "", "", "", "", fmt.Errorf("invalid asset URI %q", raw)
	}
	if strings.ContainsAny(parts[1], "\\\x00") || strings.HasPrefix(parts[1], "/") {
		return "", "", "", "", fmt.Errorf("invalid asset path in URI %q", raw)
	}
	for _, component := range strings.Split(parts[1], "/") {
		if component == ".." || component == "." {
			return "", "", "", "", fmt.Errorf("invalid asset path in URI %q", raw)
		}
	}
	constraint, err := queryValue(parsed, "version")
	if err != nil {
		return "", "", "", "", err
	}
	return parsed.Host, constraint, parts[0], filepath.ToSlash(parts[0] + "/" + parts[1]), nil
}

func queryValue(parsed *url.URL, key string) (string, error) {
	values, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", fmt.Errorf("parse URI query: %w", err)
	}
	return values.Get(key), nil
}
