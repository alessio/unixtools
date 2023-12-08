package path

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var ListSeparator = string(os.PathListSeparator)

type dirList struct {
	lst []string
}

func newDirList(lst []string) *dirList {
	return &dirList{lst: lst}
}

func NewPathList(v string) List {
	return newDirList(makePathList(os.Getenv(v)))
}

func (p *dirList) String() interface{} {
	return strings.Join(p.lst, ListSeparator)
}

func (p *dirList) StringSlice() []string {
	return p.lst
}

func (p *dirList) Prepend(path string) bool {
	cleanPath := normalizePath(path)
	if idx := slices.Index(p.lst, cleanPath); idx == -1 {
		p.lst = append([]string{cleanPath}, p.lst...)
		return true
	}

	return false
}

func (p *dirList) Append(path string) bool {
	cleanPath := normalizePath(path)
	if idx := slices.Index(p.lst, cleanPath); idx == -1 {
		p.lst = append(p.lst, cleanPath)
		return true
	}

	return false
}

func (p *dirList) Drop(path string) bool {
	cleanPath := normalizePath(path)
	if idx := slices.Index(p.lst, cleanPath); idx != -1 {
		p.lst = slices.Delete(p.lst, idx, idx+1)
		return true
	}

	return false
}

func makePathList(pathStr string) []string {
	if pathStr == "" {
		return nil
	}

	rawList := strings.Split(pathStr, ListSeparator)
	cleanList := make([]string, len(rawList))

	for i, s := range rawList {
		cleanList[i] = normalizePath(s)
	}

	return cleanList
}

func normalizePath(s string) string {
	return filepath.Clean(s)
}
