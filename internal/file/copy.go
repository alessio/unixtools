package file

/*
This file was downloaded from https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04

The original source code was published under the terms of the following
license:

	Copyright (c) 2017 Roland Singer [roland.singer@desertbit.com]

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	if _, err = os.Stat(dst); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		return fmt.Errorf("destination already exists")
	}

	if err := os.MkdirAll(dst, si.Mode()); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}

			continue
		}

		// copy symlinks without following them
		if entry.Mode()&os.ModeSymlink != 0 {
			// the original source code was skipping symbolic links
			if err := copySymlink(srcPath, dstPath); err != nil {
				return err
			}

			continue
		}

		if err := copyRegular(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// CopyRegular copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage. Unlike CopySymlink, it
// copies the symbolic link's target  instead of the symbolic link itself.
func copyRegular(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}

	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// copySymlink copies a symbolic by replicating the contents of the original
// src symbolic link. The file will be created if it does not already exist. If the
// destination file exists, it will be overwritten.
func copySymlink(src string, dst string) error {
	in, err := os.Readlink(src)
	if err != nil {
		return err
	}

	return os.Symlink(in, dst)
}
