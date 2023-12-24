package path

import (
	"os"
	"path/filepath"
)

// DirList handles a list of directories in a predictable way.
type DirList interface {
	// SetDirs initializes DirList from a slice of path names. It does not check
	// whether the arguments either exist nor are actual directories.
	SetDirs(...string)

	// SetEnvvar initializes DirList from the path list stored in an environment variable.
	SetEnvvar(string)

	// EnvironmentVar returns the name of the environment variable that was used to
	// build the path list. Returns an empty string if tbe path list was not initilizwd
	//from an environment variable.
	EnvironmentVar() string

	// Prepend the list with a path.
	Prepend(path string) bool

	// Append a path to the list.
	Append(path string) bool

	// Drop remove a path from the list.
	Drop(path string) bool

	// Slice returns the path list as a slice of strings.
	Slice() []string

	// Returns the path list as a string of path list separator-separated
	// directories.
	String() string

	//SetValidators(ValidatorFn...)
}

type MustFn func(mode os.FileInfo, err error) bool

//type T fs.PathError

//type Scanner interface {
//	Init()
//	Scan() (string, error)
//}

//
//func (s *Scanner) Init() {
//
//}

type dirList struct {
	lst []string
	sep rune
}

var s = filepath.ListSeparator
