package path

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alessio/shellescape"
)

type pathLst struct {
	dirs   []string
	envvar string
}

func (p *pathLst) setDirs(dirs ...string) {
	if len(dirs) == 0 {
		return
	}

	var cleanDirs []string
	for _, d := range dirs {
		cleanDirs = append(cleanDirs, filepath.Clean(d))
	}

	p.dirs = cleanDirs
}

func NewDirList(dirs ...string) DirList {
	return &pathLst{dirs: dirs}
}

func (p *pathLst) EnvironmentVar() string {
	return p.envvar
}

func (p *pathLst) SetEnvvar(varname string) {
	p.dirs = filepath.SplitList(
		strings.Trim(os.Getenv(varname), string(filepath.ListSeparator)))
	p.envvar = varname
}

func (p *pathLst) SetDirs(dirs ...string) {
	p.dirs = dirs
}

//func (p *pathLst) Parse(v string) { p.dirs = p.makePathList(v) }

func (p *pathLst) String() string {
	var b strings.Builder
	for _, s := range p.dirs {
		b.WriteString(shellescape.Quote(s))
		b.WriteRune(filepath.ListSeparator)
	}
	return b.String()[:len(b.String())-1]
	//	return strings.Join(p.dirs, ListSeparator)
}

func (p *pathLst) StringSlice() []string {
	return p.dirs
}

func (p *pathLst) Prepend(path string) bool {
	cleanPath := filepath.Clean(path)
	if idx := slices.Index(p.dirs, cleanPath); idx == -1 {
		p.dirs = append([]string{cleanPath}, p.dirs...)
		return true
	}

	return false
}

func (p *pathLst) Append(path string) bool {
	cleanPath := filepath.Clean(path)
	if idx := slices.Index(p.dirs, cleanPath); idx == -1 {
		p.dirs = append(p.dirs, cleanPath)
		return true
	}

	return false
}

func (p *pathLst) Drop(path string) bool {
	cleanPath := filepath.Clean(path)
	if idx := slices.Index(p.dirs, cleanPath); idx != -1 {
		p.dirs = slices.Delete(p.dirs, idx, idx+1)
		return true
	}

	return false
}

func (p *pathLst) Slice() []string { return p.dirs }

//func (p *pathLst) makePathList(pathStr string) []string {
//	if pathStr == "" {
//		return nil
//	}
//
//	rawList := strings.Split(pathStr, string(filepath.ListSeparator))
//	cleanList := make([]string, len(rawList))
//
//	for i, s := range rawList {
//		cleanList[i] = filepath.Clean(s)
//	}
//
//	return cleanList
//}
