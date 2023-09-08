package path

import (
	"path"
	"strings"
)

func PushDirIfNotInPath(path string, s string) string {
	s = normalizePath(s)
	if path == "" {
		return s
	}

	for _, d := range strings.Split(path, ":") {
		if s == normalizePath(d) {
			return path
		}
	}

	return strings.Join([]string{s, path}, ":")
}

func AddDir(path string, s string, append bool) string {
	s = normalizePath(s)
	if path == "" {
		return s
	}

	for _, d := range strings.Split(path, ":") {
		if s == normalizePath(d) {
			return path
		}
	}

	if append {
		return strings.Join([]string{path, s}, ":")
	}

	return strings.Join([]string{s, path}, ":")
}

func RemoveDir(path string, s string) string {
	s = normalizePath(s)
	if s == "" || path == "" {
		return path
	}

	var newPath []string
	for _, d := range strings.Split(path, ":") {
		if s == normalizePath(d) {
			continue
		}

		newPath = append(newPath, d)
	}

	return strings.Join(newPath, ":")
}

func normalizePath(s string) string {
	return path.Clean(strings.TrimRight(s, "/"))
}
