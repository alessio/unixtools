// Package dirlist implements functions to manipulate PATH-like
// environment variables.
package dirlist

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alessio/shellescape"
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
	//	Nil() bool

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
	return slices.Contains(d.lst, quoteAndClean(p))
}

func (d *dirList) Reset() {
	d.init()
}

// func (d *dirList) Nil() bool {
// 	return d.l == nil || len(d.l) == 0
// }

func (d *dirList) Load(s string) {
	d.src = s
	d.load()
}

func (d *dirList) LoadEnv(s string) {
	d.Load(os.Getenv(s))
}

func (d *dirList) Slice() (dst []string) {
	if len(d.lst) == 0 {
		return
	}

	dst = make([]string, len(d.lst))
	if n := copy(dst, d.lst); n == len(d.lst) {
		return dst
	}

	panic("couldn't copy the list")
}

func (d *dirList) String() string {
	if len(d.lst) == 0 {
		return ""
	}

	return strings.Join(d.lst, string(filepath.ListSeparator))
}

func (d *dirList) load() {
	d.lst = d.cleanPathVar()
}

func (d *dirList) Append(path string) {
	p := quoteAndClean(path)
	if len(d.lst) == 0 {
		d.lst = []string{p}
		return
	}

	if !d.Contains(p) {
		d.lst = append(d.lst, p)
	}
}

func (d *dirList) Drop(path string) {
	if len(d.lst) == 0 {
		return
	}

	p := quoteAndClean(path)

	if idx := slices.Index(d.lst, p); idx != -1 {
		d.lst = slices.Delete(d.lst, idx, idx+1)
	}
}

func (d *dirList) Prepend(path string) {
	p := quoteAndClean(path)
	if len(d.lst) == 0 {
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
	return cleanPathVar(d.src)
}

func cleanPathVar(src string) []string {
	if src == "" {
		return nil
	}

	pthSlice := filepath.SplitList(src)
	if len(pthSlice) == 0 {
		return nil
	}

	return removeDups(pthSlice, filterEmptyStrings)
}

func (d *dirList) clone(o *dirList) *dirList {
	o.src = d.src

	n := len(d.lst)
	o.lst = make([]string, n)

	if m := copy(o.lst, d.lst); n != m {
		panic(fmt.Sprintf("copy: expected %d items, got %d", n, m))
	}

	return o
}

func removeDups(col []string, applyFn func(string) (string, bool)) []string {
	var uniq = make([]string, 0)
	ks := make(map[string]interface{})

	for _, el := range col {
		vv, ok := applyFn(el)
		if !ok {
			continue
		}

		if _, ok := ks[vv]; !ok {
			quoted := shellescape.Quote(vv)
			uniq = append(uniq, quoted)
			ks[vv] = struct{}{}
		}
	}

	return uniq
}

var filterEmptyStrings = func(s string) (string, bool) {
	if strings.TrimSpace(s) == "" {
		return s, false
	}

	// I removed the following because filepath.Clean()
	// never returns "".
	//
	// clean := filepath.Clean(s)
	// if clean == "" {
	// 	return clean, false
	// }

	return filepath.Clean(s), true
}

func quoteAndClean(s string) string {
	return shellescape.Quote(filepath.Clean(s))
}
