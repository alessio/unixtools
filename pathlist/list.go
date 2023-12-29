// Package pathlist implements functions to manipulate PATH-like
// environment variables.
package pathlist

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// List builds a list of directories by parsing PATH-like variables
// and can perform operations on it such as append, prepend, or remove,
// while keeping the list duplicate-free.
type List interface {
	// Reset resets the list of directories to an empty slice.
	Reset()

	// Contains returns true if the list contains the path.
	Contains(string) bool

	// Nil returns true if the list is emppty.
	Nil() bool

	// Load reads the list of directories from a string.
	Load(string)

	// LoadEnv parses the value of an environment variable. It expects
	// the value to be a string of directories separates by the rune
	// filepath.ListSeparator.
	LoadEnv(string)

	// Prepend the list with a path.
	Prepend(string)

	// Append a path to the list.
	Append(string)

	// Drop remove a path from the list.
	Drop(string)

	// Slice returns the path list as a slice of strings.
	Slice() []string

	// String returns the path list as a string of path list
	// separator-separated directories.
	String() string
}

type dirList struct {
	lst []string
	src string
}

// New creates a new path list.
func New() List {
	d := new(dirList)
	d.init()
	return d
}

func (d *dirList) Contains(p string) bool {
	return slices.Contains(d.lst, p)
}

func (d *dirList) Reset() {
	d.init()
}

func (d *dirList) Nil() bool {
	return d.lst == nil || len(d.lst) == 0
}

func (d *dirList) Load(s string) {
	d.src = s
	d.load()
}

func (d *dirList) LoadEnv(s string) {
	d.Load(os.Getenv(s))
}

func (d *dirList) Slice() []string {
	if d.Nil() {
		return []string{}
	}

	dst := make([]string, len(d.lst))
	n := copy(dst, d.lst)
	if n != len(d.lst) {
		panic("couldn't copy the list")
	}

	return dst
}

func (d *dirList) String() string {
	if !d.Nil() {
		return strings.Join(d.lst, string(filepath.ListSeparator))
	}

	return ""
}

func (d *dirList) load() {
	d.lst = d.cleanPathVar()
}

func (d *dirList) Append(path string) {
	p := filepath.Clean(path)
	if d.Nil() {
		d.lst = []string{p}
		return
	}

	if !d.Contains(p) {
		d.lst = append(d.lst, p)
	}
}

func (d *dirList) Drop(path string) {
	if d.Nil() {
		return
	}
	p := filepath.Clean(path)

	if idx := slices.Index(d.lst, p); idx != -1 {
		d.lst = slices.Delete(d.lst, idx, idx+1)
	}
}

func (d *dirList) Prepend(path string) {
	p := filepath.Clean(path)
	if d.Nil() {
		d.lst = []string{p}
		return
	}

	if !d.Contains(p) {
		d.lst = slices.Insert(d.lst, 0, p)
	}
}

func (d *dirList) init() {
	d.src = ""
	d.lst = []string{}
}

func (d *dirList) cleanPathVar() []string {
	if d.src == "" {
		return nil
	}

	pthSlice := filepath.SplitList(d.src)
	if pthSlice == nil {
		return nil
	}

	return removeDups(pthSlice, filterEmptyStrings)
}

func (d *dirList) clone(o *dirList) *dirList {
	o.src = d.src
	o.lst = make([]string, len(d.lst))
	copy(o.lst, d.lst)

	return o
}

func removeDups[T comparable](col []T, applyFn func(T) (T, bool)) []T {
	var uniq = make([]T, 0)
	ks := make(map[T]interface{})

	for _, el := range col {
		vv, ok := applyFn(el)
		if !ok {
			continue
		}

		if _, ok := ks[vv]; !ok {
			uniq = append(uniq, vv)
			ks[vv] = struct{}{}
		}
	}

	return uniq
}

var filterEmptyStrings = func(s string) (string, bool) {
	clean := filepath.Clean(s)
	if clean != "" {
		return clean, true
	}

	return clean, false
}
