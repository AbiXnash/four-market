package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

var ErrPathTraversal = fmt.Errorf("path traversal detected")

func SafeJoin(base, target string) (string, error) {
	base, err := filepath.Abs(filepath.Clean(base))
	if err != nil {
		return "", err
	}

	target = filepath.Clean(target)
	if filepath.IsAbs(target) {
		return "", ErrPathTraversal
	}

	joined := filepath.Join(base, target)
	if !strings.HasPrefix(joined, base) {
		return "", ErrPathTraversal
	}

	return joined, nil
}
