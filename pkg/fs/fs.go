package fs

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
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

func OpenFile(filename string) (*os.File, fs.FileInfo, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error file %s not open: %w", filename, err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("error to get file size: %w", err)
	}
	return f, stat, nil
}

func MustFilesInDot(suffix ...string) []string {
	if files, err := FilesInDot(suffix...); err == nil {
		return files
	}
	return []string{}
}

func MustEntitiesInDot(suffix ...string) []string {
	if ent, err := EntitiesInDot(suffix...); err == nil {
		return ent
	}
	return []string{}
}

func FilesInDot(suffix ...string) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	return filesInDir(wd, false, suffix...)
}

func EntitiesInDot(suffix ...string) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	return filesInDir(wd, true, suffix...)
}

func filesInDir(dir string, includeDirs bool, suffix ...string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory entries: %w", err)
	}

	for _, entry := range entries {
		if !includeDirs && entry.IsDir() {
			continue
		}
		if matchesSuffix(entry.Name(), suffix) {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
func matchesSuffix(name string, suffix []string) bool {
	if len(suffix) == 0 {
		return true
	}
	for _, suf := range suffix {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}
