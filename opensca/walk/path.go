package walk

import (
	"fmt"
	"path/filepath"
	"strings"
)

func resolveExtractPath(base, entry string) (string, error) {
	normalized := strings.ReplaceAll(entry, "\\", string(filepath.Separator))
	cleaned := filepath.Clean(normalized)
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("invalid archive entry path %q", entry)
	}

	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("archive entry path %q is absolute", entry)
	}

	target := filepath.Join(base, cleaned)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry path %q escapes target directory", entry)
	}

	return target, nil
}
