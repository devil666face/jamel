package fs

import (
	"fmt"
	"io/fs"
	"os"
)

func WriteFile(
	path string,
	data []byte,
	perm ...int,
) error {
	if len(perm) == 0 {
		perm = append(perm, 0644)
	}
	if err := os.WriteFile(path, data, fs.FileMode(perm[0])); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
