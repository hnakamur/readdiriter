// Package readdiriter provides functions which return an iterator
// over directory entries in the specified directory.
package readdiriter

import (
	"errors"
	"io"
	"io/fs"
	"iter"
)

// NewReadDirIter returns an iterate over directory entries from the file
// parameter.
// The n parameter follows the semantics of fs.ReadDirFile:
// https://pkg.go.dev/io/fs@latest#ReadDirFile.
//
// Note: The directory entries are not in lexical order.
func NewReadDirIter(file fs.ReadDirFile, n int) iter.Seq2[fs.DirEntry, error] {
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
