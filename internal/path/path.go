package path

import (
	"path/filepath"
	"strings"
)

func PushDirIfNotInPath(path string, s string) string {
	return AddDir(path, s, false)
}

func AddDir(path string, s string, append bool) string {
	s = filepath.Clean(s)
	if path == "" {
		return s
	}

	for _, d := range strings.Split(path, ":") {
		if s == filepath.Clean(d) {
			return path
		}
	}

	if append {
		return strings.Join([]string{path, s}, ":")
	}

	return strings.Join([]string{s, path}, ":")
}

func RemoveDir(path string, s string) string {
	s = filepath.Clean(s)
	if s == "" || path == "" {
		return path
	}

	newPath := make([]string, 0)

	for _, d := range strings.Split(path, ":") {
		if s == filepath.Clean(d) {
			continue
		}

		newPath = append(newPath, d)
	}

	return strings.Join(newPath, ":")
}
