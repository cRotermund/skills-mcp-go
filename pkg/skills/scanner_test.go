package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScannerLoadsValidSkillsAndTextAssets(t *testing.T) {
	root := t.TempDir()
	writeSkillFixture(t, root, "nested/demo", "1.2.0", "Demo skill")
	assetDir := filepath.Join(root, "nested", "demo", "scripts", "helpers")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "check.sh"), []byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := NewScanner().Scan(root)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("loaded %d skills, want 1", len(result.Skills))
	}
	if len(result.Skills[0].Assets) != 1 || result.Skills[0].Assets[0].RelativePath != "scripts/helpers/check.sh" {
		t.Fatalf("assets = %#v", result.Skills[0].Assets)
	}
}

func TestScannerSkipsMalformedSkillsAndUnsupportedAssets(t *testing.T) {
	root := t.TempDir()
	writeSkillFixture(t, root, "valid", "1.0.0", "Valid")
	invalidDir := filepath.Join(root, "invalid")
	if err := os.MkdirAll(invalidDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(invalidDir, "SKILL.md"), []byte("not frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}
	assets := filepath.Join(root, "valid", "tools")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "binary.dat"), []byte{0, 1, 2}, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := NewScanner().Scan(root)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 1 || result.Skills[0].Name() != "valid" {
		t.Fatalf("skills = %#v", result.Skills)
	}
	if len(result.Diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestScannerSkipsDuplicateVersions(t *testing.T) {
	root := t.TempDir()
	writeSkillFixture(t, root, "one", "1.0.0", "One")
	writeSkillFixture(t, root, "two", "1.0.0", "Two")
	first, err := os.ReadFile(filepath.Join(root, "one", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	first = []byte(strings.Replace(string(first), "name: one", "name: same", 1))
	if err := os.WriteFile(filepath.Join(root, "one", "SKILL.md"), first, 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(root, "two", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	second = []byte(strings.Replace(string(second), "name: two", "name: same", 1))
	if err := os.WriteFile(filepath.Join(root, "two", "SKILL.md"), second, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := NewScanner().Scan(root)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 1 || len(result.Diagnostics) != 1 {
		t.Fatalf("skills=%d diagnostics=%d", len(result.Skills), len(result.Diagnostics))
	}
}

func writeSkillFixture(t *testing.T, root, relative, version, description string) {
	t.Helper()
	directory := filepath.Join(root, relative)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(directory)
	content := "---\nname: " + name + "\nversion: " + version + "\ndescription: " + description + "\n---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(directory, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
