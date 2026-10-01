package skills

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func readAsset(skill Skill, relativePath string) (Asset, error) {
	if relativePath == "" || filepath.IsAbs(relativePath) {
		return Asset{}, fmt.Errorf("asset path must be relative")
	}
	relativePath = filepath.Clean(filepath.FromSlash(relativePath))
	if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return Asset{}, fmt.Errorf("asset path escapes skill directory")
	}
	parts := strings.Split(relativePath, string(os.PathSeparator))
	if len(parts) < 2 || (parts[0] != AssetScripts && parts[0] != AssetTools) {
		return Asset{}, fmt.Errorf("asset must be below scripts or tools")
	}

	current := skill.Directory
	for _, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return Asset{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return Asset{}, fmt.Errorf("asset path contains symlink")
		}
	}
	info, err := os.Stat(current)
	if err != nil {
		return Asset{}, err
	}
	if !info.Mode().IsRegular() {
		return Asset{}, fmt.Errorf("asset is not a regular file")
	}
	if info.Size() > DefaultMaxAssetBytes {
		return Asset{}, fmt.Errorf("asset size exceeds limit")
	}
	data, err := os.ReadFile(current)
	if err != nil {
		return Asset{}, err
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return Asset{}, fmt.Errorf("asset is not readable text")
	}
	for _, asset := range skill.Assets {
		if filepath.Clean(filepath.FromSlash(asset.RelativePath)) == relativePath {
			return asset, nil
		}
	}
	return Asset{}, fmt.Errorf("asset %q not found", relativePath)
}

func ReadAssetText(skill Skill, relativePath string) (Asset, string, error) {
	asset, err := readAsset(skill, relativePath)
	if err != nil {
		return Asset{}, "", err
	}
	data, err := os.ReadFile(filepath.Join(skill.Directory, filepath.FromSlash(relativePath)))
	if err != nil {
		return Asset{}, "", err
	}
	return asset, string(data), nil
}
