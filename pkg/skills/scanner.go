package skills

import (
	"bytes"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Masterminds/semver/v3"
)

type Scanner struct {
	MaxAssetBytes int64
}

func NewScanner() Scanner {
	return Scanner{MaxAssetBytes: DefaultMaxAssetBytes}
}

func (s Scanner) Scan(root string) (ScanResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ScanResult{}, fmt.Errorf("resolve skill root: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return ScanResult{}, fmt.Errorf("stat skill root: %w", err)
	}
	if !info.IsDir() {
		return ScanResult{}, fmt.Errorf("skill root %q is not a directory", root)
	}

	result := ScanResult{}
	seen := make(map[string]string)
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: path, Message: "walk path", Err: walkErr})
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() && path != root && (entry.Name() == AssetScripts || entry.Name() == AssetTools) {
			return filepath.SkipDir
		}
		if entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}

		skill, parseErr := ParseSkillFile(path)
		if parseErr != nil {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: path, Message: "skip invalid skill", Err: parseErr})
			return nil
		}
		version, versionErr := semver.NewVersion(skill.Version())
		if versionErr != nil {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: path, Message: "skip invalid skill version", Err: versionErr})
			return nil
		}
		key := skill.Name() + "@" + version.String()
		if previous, exists := seen[key]; exists {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Path: path, Message: "skip duplicate skill version", Err: fmt.Errorf("already loaded from %s", previous)})
			return nil
		}
		assets, diagnostics := s.scanAssets(skill.Directory)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		skill.Assets = assets
		seen[key] = path
		result.Skills = append(result.Skills, skill)
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("scan skill root: %w", err)
	}
	sort.Slice(result.Skills, func(i, j int) bool {
		if result.Skills[i].Name() == result.Skills[j].Name() {
			return result.Skills[i].Version() < result.Skills[j].Version()
		}
		return result.Skills[i].Name() < result.Skills[j].Name()
	})
	return result, nil
}

func (s Scanner) scanAssets(skillDir string) ([]Asset, []Diagnostic) {
	assets := []Asset{}
	diagnostics := []Diagnostic{}
	for _, kind := range []string{AssetScripts, AssetTools} {
		root := filepath.Join(skillDir, kind)
		info, err := os.Stat(root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Path: root, Message: "inspect asset directory", Err: err})
			continue
		}
		if !info.IsDir() {
			diagnostics = append(diagnostics, Diagnostic{Path: root, Message: "skip asset path that is not a directory"})
			continue
		}
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				diagnostics = append(diagnostics, Diagnostic{Path: path, Message: "inspect asset", Err: walkErr})
				return nil
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				diagnostics = append(diagnostics, Diagnostic{Path: path, Message: "skip symlink asset"})
				return nil
			}
			asset, ok, assetErr := classifyAsset(path, kind, s.MaxAssetBytes)
			if assetErr != nil {
				diagnostics = append(diagnostics, Diagnostic{Path: path, Message: "skip unsupported asset", Err: assetErr})
				return nil
			}
			if ok {
				relative, relErr := filepath.Rel(skillDir, path)
				if relErr != nil {
					diagnostics = append(diagnostics, Diagnostic{Path: path, Message: "resolve asset path", Err: relErr})
					return nil
				}
				asset.RelativePath = filepath.ToSlash(relative)
				assets = append(assets, asset)
			}
			return nil
		})
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].RelativePath < assets[j].RelativePath })
	return assets, diagnostics
}

func classifyAsset(path, kind string, maxBytes int64) (Asset, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Asset{}, false, err
	}
	if !info.Mode().IsRegular() {
		return Asset{}, false, fmt.Errorf("not a regular file")
	}
	if info.Size() > maxBytes {
		return Asset{}, false, fmt.Errorf("size %d exceeds limit %d", info.Size(), maxBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Asset{}, false, err
	}
	if !utf8.Valid(data) {
		return Asset{}, false, fmt.Errorf("invalid UTF-8")
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return Asset{}, false, fmt.Errorf("contains NUL byte")
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if mimeType == "" || !strings.HasPrefix(mimeType, "text/") && mimeType != "application/json" && mimeType != "application/yaml" {
		mimeType = "text/plain"
	}
	return Asset{Kind: kind, MIMEType: mimeType, Size: info.Size()}, true, nil
}
