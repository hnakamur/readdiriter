package readdiriter

import (
	"errors"
	"io"
	"io/fs"
	"iter"
)

func NewReadDirIter(file fs.ReadDirFile, n int) iter.Seq2[fs.DirEntry, error] {
	return func(yield func(fs.DirEntry, error) bool) {
		if n <= 0 {
			de, err := file.ReadDir(n)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, e := range de {
				if !yield(e, nil) {
					return
				}
			}
			return
		}

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
			if seenEOF {
				return
			}
		}
	}
}
