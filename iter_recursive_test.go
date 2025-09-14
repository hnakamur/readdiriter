package readdiriter

import (
	"flag"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestNewReadDirIterRecursive(t *testing.T) {
	dir := tempDir(t)
	t.Logf("dir=%s", dir)

	inputDirs := [][]string{
		{"a"},
		{"a", "b"},
		{"a", "b", "c"},
		{"a", "d"},
	}
	inputFiles := [][]string{
		{"a", "f1"},
		{"a", "f2"},
		{"a", "b", "c", "f2"},
	}

	wantDirs := make([]string, len(inputDirs))
	for i, inputDir := range inputDirs {
		dirPath := filepath.Join(append([]string{dir}, inputDir...)...)
		if err := os.Mkdir(dirPath, 0o700); err != nil {
			t.Fatal(err)
		}
		t.Logf("cretaed dir=%s", dirPath)
		wantDirs[i] = dirPath
	}
	slices.Sort(wantDirs)

	wantFiles := make([]string, len(inputFiles))
	for i, inputFile := range inputFiles {
		filePath := filepath.Join(append([]string{dir}, inputFile...)...)
		if err := os.WriteFile(filePath, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Logf("cretaed file=%s", filePath)
		wantFiles[i] = filePath
	}
	slices.Sort(wantFiles)

	dirFile, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer dirFile.Close()

	openDir := func(name string) (fs.ReadDirFile, error) {
		return os.Open(name)
	}

	var gotDirs, gotFiles []string
	for de, err := range NewReadDirIterRecursive(dir, dirFile, 0, openDir, nil) {
		if err != nil {
			t.Fatal(err)
		}
		relPath, err := filepath.Rel(dir, de.Name())
		if err != nil {
			t.Fatal(err)
		}
		slog.Info("NewReadDirIterRecursive loop", "de.Name()", de.Name(), "isDir", de.IsDir(), "relPath", relPath)
		if de.IsDir() {
			gotDirs = append(gotDirs, de.Name())
		} else {
			gotFiles = append(gotFiles, de.Name())
		}
	}
	slices.Sort(gotDirs)
	slices.Sort(gotFiles)

	if !slices.Equal(gotDirs, wantDirs) {
		t.Errorf("dirs mismatch,\n got=%v,\nwant=%v", gotDirs, wantDirs)
	}
	if !slices.Equal(gotFiles, wantFiles) {
		t.Errorf("files mismatch,\n got=%v,\nwant=%v", gotFiles, wantFiles)
	}
}

var keepTempDir = flag.Bool("keep-temp-dir", false, "Whether to keep temporary directory")

func tempDir(t *testing.T) string {
	t.Helper()

	if *keepTempDir {
		dir, err := os.MkdirTemp("", t.Name()+"-")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("temp-dir=%s", dir)
		return dir
	}
	return t.TempDir()
}
