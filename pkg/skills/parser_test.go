package skills

import (
	"strings"
	"testing"
)

func TestParseSkillPreservesRawMarkdown(t *testing.T) {
	input := "---\nname: demo\nversion: 1.2.0\ndescription: Demo skill\ntriggers:\n  - demo\nallowed_tools:\n  - read_file\n---\n\n# Demo\n\nInstructions.\n"

	skill, err := ParseSkill([]byte(input), "/skills/demo/SKILL.md")
	if err != nil {
		t.Fatalf("ParseSkill() error = %v", err)
	}
	if skill.RawMarkdown != input {
		t.Fatalf("RawMarkdown was changed")
	}
	if skill.Body != "\n# Demo\n\nInstructions.\n" {
		t.Fatalf("Body = %q", skill.Body)
	}
	if skill.Name() != "demo" || skill.Version() != "1.2.0" {
		t.Fatalf("metadata = %s@%s", skill.Name(), skill.Version())
	}
	if len(skill.Frontmatter.Triggers) != 1 || len(skill.Frontmatter.AllowedTools) != 1 {
		t.Fatalf("list metadata was not parsed: %#v", skill.Frontmatter)
	}
}

func TestParseSkillAcceptsCRLF(t *testing.T) {
	input := "---\r\nname: demo\r\nversion: 1.0.0\r\ndescription: Demo\r\n---\r\nbody\r\n"
	skill, err := ParseSkill([]byte(input), "SKILL.md")
	if err != nil {
		t.Fatalf("ParseSkill() error = %v", err)
	}
	if skill.Body != "body\r\n" {
		t.Fatalf("Body = %q", skill.Body)
	}
}

func TestParseSkillRejectsMalformedMetadata(t *testing.T) {
	tests := map[string]string{
		"missing delimiter": "name: demo\n",
		"unterminated":      "---\nname: demo\nversion: 1.0.0\ndescription: Demo\n",
		"missing name":      "---\nversion: 1.0.0\ndescription: Demo\n---\n",
		"bad version":       "---\nname: demo\nversion: nope\ndescription: Demo\n---\n",
		"unknown field":     "---\nname: demo\nversion: 1.0.0\ndescription: Demo\nunknown: true\n---\n",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseSkill([]byte(input), "SKILL.md"); err == nil {
				t.Fatal("ParseSkill() succeeded unexpectedly")
			}
		})
	}
}

func TestParseSkillDefaultsOptionalLists(t *testing.T) {
	skill, err := ParseSkill([]byte("---\nname: demo\nversion: 1.0.0\ndescription: Demo\n---\nbody"), "SKILL.md")
	if err != nil {
		t.Fatalf("ParseSkill() error = %v", err)
	}
	if skill.Frontmatter.Triggers == nil || skill.Frontmatter.AllowedTools == nil {
		t.Fatal("optional lists should be non-nil")
	}
	if strings.TrimSpace(skill.Body) != "body" {
		t.Fatalf("Body = %q", skill.Body)
	}
}
