package readdiriter

import (
	"io/fs"
	"iter"
	"path/filepath"
)

// ReadDirCloser is the interface that groups ReadDir and Close methods.
type ReadDirCloser interface {
	ReadDirer
	Close() error
}

// OpenReadDirCloserFunc is the type of the function called by NewReadDirIterRecursive
// to open each directory.
type OpenReadDirCloserFunc func(name string) (ReadDirCloser, error)

// DirAndEntry is the struct that groups the directory and io/fs.DirEntry.
type DirAndEntry struct {
	dir   string
	entry fs.DirEntry
}

// Dir returns the directory in a DirAndEntry struct.
func (e *DirAndEntry) Dir() string {
	return e.dir
}

// Entry returns the directory entry in a DirAndEntry struct.
func (e *DirAndEntry) Entry() fs.DirEntry {
	return e.entry
}

// NewReadDirIterRecursive returns an iterate over directory entries by walking
// each directory in the tree, including baseDir.
//
// The n parameter follows the semantics of fs.ReadDirFile:
// https://pkg.go.dev/io/fs@latest#ReadDirFile.
//
// Note: The directory entries are not in lexical order in each directory.
func NewReadDirIterRecursive(baseDir string, openDir OpenReadDirCloserFunc, n int, skipDir *bool) iter.Seq2[*DirAndEntry, error] {
	return func(yield func(*DirAndEntry, error) bool) {
		walkDir(baseDir, openDir, n, skipDir, yield)
	}
}

func walkDir(baseDir string, openDir OpenReadDirCloserFunc, n int, skipDir *bool, yield func(*DirAndEntry, error) bool) {
	dirFile, err := openDir(baseDir)
	if err != nil {
		yield(nil, err)
		return
	}
	defer func() {
		if err := dirFile.Close(); err != nil {
			yield(nil, err)
			return
		}
	}()

	for entry, err := range NewReadDirIter(dirFile, n) {
		if err != nil {
			yield(nil, err)
			return
		}
		if !yield(&DirAndEntry{
			dir:   baseDir,
			entry: entry,
		}, nil) {
			return
		}
		if entry.IsDir() {
			if skipDir != nil && *skipDir {
				*skipDir = false
				continue
			}

			subDir := filepath.Join(baseDir, entry.Name())
			walkDir(subDir, openDir, n, skipDir, yield)
		}
	}
}
