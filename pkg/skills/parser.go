package skills

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

func ParseSkillFile(path string) (Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}
	return ParseSkill(data, path)
}

func ParseSkill(data []byte, path string) (Skill, error) {
	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return Skill{}, err
	}

	var metadata Frontmatter
	decoder := yaml.NewDecoder(bytes.NewReader(frontmatter))
	decoder.KnownFields(true)
	if err := decoder.Decode(&metadata); err != nil {
		return Skill{}, fmt.Errorf("decode frontmatter: %w", err)
	}
	if strings.TrimSpace(metadata.Name) == "" {
		return Skill{}, fmt.Errorf("frontmatter field %q is required", "name")
	}
	if strings.TrimSpace(metadata.Version) == "" {
		return Skill{}, fmt.Errorf("frontmatter field %q is required", "version")
	}
	if _, err := semver.NewVersion(metadata.Version); err != nil {
		return Skill{}, fmt.Errorf("invalid semantic version %q: %w", metadata.Version, err)
	}
	if strings.TrimSpace(metadata.Description) == "" {
		return Skill{}, fmt.Errorf("frontmatter field %q is required", "description")
	}
	if metadata.Triggers == nil {
		metadata.Triggers = []string{}
	}
	if metadata.AllowedTools == nil {
		metadata.AllowedTools = []string{}
	}

	return Skill{
		Frontmatter: metadata,
		Path:        path,
		Directory:   directoryForSkill(path),
		RawMarkdown: string(data),
		Body:        string(body),
		Assets:      []Asset{},
	}, nil
}

func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	firstEnd := lineEnd(data, 0)
	if firstEnd < 0 || strings.TrimSuffix(string(data[:firstEnd]), "\r") != "---" {
		return nil, nil, fmt.Errorf("SKILL.md must begin with YAML frontmatter delimiter")
	}

	frontStart := firstEnd
	for frontStart < len(data) && (data[frontStart] == '\r' || data[frontStart] == '\n') {
		frontStart++
	}
	for lineStart := frontStart; lineStart <= len(data); {
		lineEndOffset := lineEnd(data, lineStart)
		if lineEndOffset < 0 {
			lineEndOffset = len(data)
		}
		line := strings.TrimSuffix(string(data[lineStart:lineEndOffset]), "\r")
		if line == "---" {
			bodyStart := lineEndOffset
			if bodyStart < len(data) && data[bodyStart] == '\r' {
				bodyStart++
			}
			if bodyStart < len(data) && data[bodyStart] == '\n' {
				bodyStart++
			}
			return data[frontStart:lineStart], data[bodyStart:], nil
		}
		if lineEndOffset == len(data) {
			break
		}
		lineStart = lineEndOffset + 1
	}

	return nil, nil, fmt.Errorf("SKILL.md frontmatter is not terminated")
}

func lineEnd(data []byte, start int) int {
	if start > len(data) {
		return -1
	}
	if index := bytes.IndexByte(data[start:], '\n'); index >= 0 {
		return start + index
	}
	if start == len(data) {
		return start
	}
	return -1
}

func directoryForSkill(path string) string {
	lastSlash := strings.LastIndexAny(path, `/\\`)
	if lastSlash < 0 {
		return "."
	}
	return path[:lastSlash]
}
