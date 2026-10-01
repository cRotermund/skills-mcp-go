package skills

import "fmt"

const (
	DefaultMaxAssetBytes int64 = 1 << 20

	AssetScripts = "scripts"
	AssetTools   = "tools"
)

type Frontmatter struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Description  string   `yaml:"description"`
	Triggers     []string `yaml:"triggers"`
	AllowedTools []string `yaml:"allowed_tools"`
}

type Asset struct {
	RelativePath string `json:"path"`
	Kind         string `json:"kind"`
	MIMEType     string `json:"mime_type"`
	Size         int64  `json:"size"`
}

type Skill struct {
	Frontmatter Frontmatter `json:"frontmatter"`
	Directory   string      `json:"-"`
	Path        string      `json:"-"`
	RawMarkdown string      `json:"-"`
	Body        string      `json:"-"`
	Assets      []Asset     `json:"assets"`
}

func (s Skill) Name() string {
	return s.Frontmatter.Name
}

func (s Skill) Version() string {
	return s.Frontmatter.Version
}

type Diagnostic struct {
	Path    string
	Message string
	Err     error
}

func (d Diagnostic) Error() string {
	if d.Err == nil {
		return fmt.Sprintf("%s: %s", d.Path, d.Message)
	}
	return fmt.Sprintf("%s: %s: %v", d.Path, d.Message, d.Err)
}

type ScanResult struct {
	Skills      []Skill
	Diagnostics []Diagnostic
}
