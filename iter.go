// Package readdiriter provides functions which return an iterator
// over directory entries in the specified directory.
package readdiriter

import (
	"cmp"
	"errors"
	"io"
	"io/fs"
	"iter"
	"slices"
)

// ReadDirer is the interface that wraps the ReadDir method in io/fs.ReadDirFile
// interface.
type ReadDirer interface {
	ReadDir(n int) ([]fs.DirEntry, error)
}

// NewReadDirIter returns an iterate over directory entries from the file
// parameter.
// The n parameter follows the semantics of fs.ReadDirFile:
// https://pkg.go.dev/io/fs@latest#ReadDirFile.
//
// Note: The directory entries are not in lexical order.
func NewReadDirIter(file ReadDirer, n int) iter.Seq2[fs.DirEntry, error] {
	return func(yield func(fs.DirEntry, error) bool) {
		for {
			de, err := file.ReadDir(n)
			var seenEOF bool
			if err != nil {
				if !errors.Is(err, io.EOF) {
					yield(nil, err)
					return
				}
				seenEOF = true
			}
			for _, e := range de {
				if !yield(e, nil) {
					return
				}
			}
			if seenEOF || (n == 0 && len(de) == 0) {
				return
			}
		}
	}
}

// NewReadDirIterSorted returns an iterate over directory entries from the file
// parameter.
//
// Note: The directory entries are in lexical order.
func NewReadDirIterSorted(file ReadDirer) iter.Seq2[fs.DirEntry, error] {
	return func(yield func(fs.DirEntry, error) bool) {
		de, err := file.ReadDir(0)
		if err != nil {
			yield(nil, err)
			return
		}
		sortDirEntriesByName(de)
		for _, e := range de {
			if !yield(e, nil) {
				return
			}
		}
	}
}

func sortDirEntriesByName(de []fs.DirEntry) {
	slices.SortFunc(de, func(a, b fs.DirEntry) int {
		return cmp.Compare(a.Name(), b.Name())
	})
}
