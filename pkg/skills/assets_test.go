package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAssetTextRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	writeSkillFixture(t, root, "demo", "1.0.0", "Demo")
	assetDir := filepath.Join(root, "demo", "scripts")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "check.sh"), []byte("echo ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := NewScanner().Scan(root)
	if err != nil || len(result.Skills) != 1 {
		t.Fatalf("Scan() skills=%d err=%v", len(result.Skills), err)
	}
	for _, path := range []string{"../SKILL.md", "/etc/passwd", "tools/other.txt"} {
		if _, _, err := ReadAssetText(result.Skills[0], path); err == nil {
			t.Errorf("ReadAssetText(%q) succeeded unexpectedly", path)
		}
	}
}
