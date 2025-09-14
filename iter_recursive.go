package readdiriter

import (
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path"
)

type OpenReadDirFileFunc func(name string) (fs.ReadDirFile, error)

var _ = (fs.ReadDirFile)((*os.File)(nil))

func NewReadDirIterRecursive(baseDir string, file fs.ReadDirFile, n int, openDir OpenReadDirFileFunc, skipDir *bool) iter.Seq2[fs.DirEntry, error] {
	return func(yield func(fs.DirEntry, error) bool) {
		walkDir(baseDir, file, n, yield, openDir, skipDir)
	}
}

func walkDir(baseDir string, file fs.ReadDirFile, n int, yield func(fs.DirEntry, error) bool, openDir OpenReadDirFileFunc, skipDir *bool) {
	slog.Info("walkDir start", "baseDir", baseDir)
	defer slog.Info("walkDir exit", "baseDir", baseDir)

	for entry, err := range NewReadDirIter(file, n) {
		if err != nil {
			yield(nil, err)
			return
		}
		slog.Info("walkDir entry", "baseDir", baseDir, "entryName", entry.Name(), "isDir", entry.IsDir())
		basedEntry := newBasedDirEntry(baseDir, entry)
		if !yield(basedEntry, nil) {
			return
		}
		if entry.IsDir() {
			if skipDir != nil && *skipDir {
				*skipDir = false
				continue
			}

			dirPath := basedEntry.Name()
			dirFile, err := openDir(dirPath)
			if err != nil {
				yield(nil, err)
				return
			}
			slog.Info("opened dir", "dirPath", dirPath)
			defer slog.Info("closed dir", "dirPath", dirPath)
			defer dirFile.Close()

			walkDir(dirPath, dirFile, n, yield, openDir, skipDir)
		}
	}
}

type basedDirEntry struct {
	base     string
	dirEntry fs.DirEntry
}

var _ = (fs.DirEntry)((*basedDirEntry)(nil))

func newBasedDirEntry(base string, dirEntry fs.DirEntry) *basedDirEntry {
	return &basedDirEntry{
		base:     base,
		dirEntry: dirEntry,
	}
}

func (e *basedDirEntry) Name() string {
	return path.Join(e.base, e.dirEntry.Name())
}

func (e *basedDirEntry) IsDir() bool {
	return e.dirEntry.IsDir()
}

func (e *basedDirEntry) Type() fs.FileMode {
	return e.dirEntry.Type()
}

func (e *basedDirEntry) Info() (fs.FileInfo, error) {
	return e.dirEntry.Info()
}
