package readdiriter

import (
	"flag"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"
)

func TestNewReadDirIterRecursive(t *testing.T) {
	t.Run("noSkip", func(t *testing.T) {
		dir := tempDir(t)

		inputDirs := []string{
			"a",
			filepath.Join("a", "b"),
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "d"),
		}
		inputFiles := []string{
			filepath.Join("a", "f1"),
			filepath.Join("a", "f2"),
			filepath.Join("a", "b", "c", "f2"),
		}

		wantDirs := make([]string, len(inputDirs))
		for i, inputDir := range inputDirs {
			dirPath := filepath.Join(dir, inputDir)
			if err := os.Mkdir(dirPath, 0o700); err != nil {
				t.Fatal(err)
			}
			wantDirs[i] = dirPath
		}
		slices.Sort(wantDirs)

		wantFiles := make([]string, len(inputFiles))
		for i, inputFile := range inputFiles {
			filePath := filepath.Join(dir, inputFile)
			if err := os.WriteFile(filePath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			wantFiles[i] = filePath
		}
		slices.Sort(wantFiles)

		openDir := func(name string) (ReadDirCloser, error) {
			return os.Open(name)
		}
		var gotDirs, gotFiles []string
		for de, err := range NewReadDirIterRecursive(dir, openDir, 0, nil) {
			if err != nil {
				t.Fatal(err)
			}
			dePath := path.Join(de.Dir(), de.Entry().Name())
			if de.Entry().IsDir() {
				gotDirs = append(gotDirs, dePath)
			} else {
				gotFiles = append(gotFiles, dePath)
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
	})
	t.Run("skip", func(t *testing.T) {
		dir := tempDir(t)

		inputDirs := []string{
			"a",
			filepath.Join("a", "b"),
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "d"),
		}
		inputFiles := []string{
			filepath.Join("a", "f1"),
			filepath.Join("a", "f2"),
			filepath.Join("a", "b", "c", "f2"),
		}

		wantDirs := []string{
			filepath.Join(dir, "a"),
			filepath.Join(dir, "a", "d"),
		}
		slices.Sort(wantDirs)
		wantFiles := []string{
			filepath.Join(dir, "a", "f1"),
			filepath.Join(dir, "a", "f2"),
		}
		slices.Sort(wantFiles)

		for _, inputDir := range inputDirs {
			dirPath := filepath.Join(dir, inputDir)
			if err := os.Mkdir(dirPath, 0o700); err != nil {
				t.Fatal(err)
			}
		}

		for _, inputFile := range inputFiles {
			filePath := filepath.Join(dir, inputFile)
			if err := os.WriteFile(filePath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}

		openDir := func(name string) (ReadDirCloser, error) {
			return os.Open(name)
		}
		skipDir := false
		var gotDirs, gotFiles []string
		for de, err := range NewReadDirIterRecursive(dir, openDir, 0, &skipDir) {
			if err != nil {
				t.Fatal(err)
			}
			dePath := path.Join(de.Dir(), de.Entry().Name())
			if de.Entry().IsDir() {
				if dePath == filepath.Join(dir, "a", "b") {
					skipDir = true
					continue
				}
				gotDirs = append(gotDirs, dePath)
			} else {
				gotFiles = append(gotFiles, dePath)
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
	})
}
func TestNewReadDirIterRecursiveSorted(t *testing.T) {
	t.Run("noSkip", func(t *testing.T) {
		dir := tempDir(t)

		inputDirs := []string{
			"a",
			filepath.Join("a", "b"),
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "d"),
		}
		inputFiles := []string{
			filepath.Join("a", "b", "c", "f2"),
			filepath.Join("a", "f1"),
			filepath.Join("a", "f2"),
		}

		wantDirs := make([]string, len(inputDirs))
		for i, inputDir := range inputDirs {
			dirPath := filepath.Join(dir, inputDir)
			if err := os.Mkdir(dirPath, 0o700); err != nil {
				t.Fatal(err)
			}
			wantDirs[i] = dirPath
		}

		wantFiles := make([]string, len(inputFiles))
		for i, inputFile := range inputFiles {
			filePath := filepath.Join(dir, inputFile)
			if err := os.WriteFile(filePath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			wantFiles[i] = filePath
		}

		openDir := func(name string) (ReadDirCloser, error) {
			return os.Open(name)
		}
		var gotDirs, gotFiles []string
		for de, err := range NewReadDirIterRecursiveSorted(dir, openDir, nil) {
			if err != nil {
				t.Fatal(err)
			}
			dePath := path.Join(de.Dir(), de.Entry().Name())
			if de.Entry().IsDir() {
				gotDirs = append(gotDirs, dePath)
			} else {
				gotFiles = append(gotFiles, dePath)
			}
		}

		if !slices.Equal(gotDirs, wantDirs) {
			t.Errorf("dirs mismatch,\n got=%v,\nwant=%v", gotDirs, wantDirs)
		}
		if !slices.Equal(gotFiles, wantFiles) {
			t.Errorf("files mismatch,\n got=%v,\nwant=%v", gotFiles, wantFiles)
		}
	})
	t.Run("skip", func(t *testing.T) {
		dir := tempDir(t)

		inputDirs := []string{
			"a",
			filepath.Join("a", "b"),
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "d"),
		}
		inputFiles := []string{
			filepath.Join("a", "b", "c", "f2"),
			filepath.Join("a", "f1"),
			filepath.Join("a", "f2"),
		}

		wantDirs := []string{
			filepath.Join(dir, "a"),
			filepath.Join(dir, "a", "d"),
		}
		wantFiles := []string{
			filepath.Join(dir, "a", "f1"),
			filepath.Join(dir, "a", "f2"),
		}

		for _, inputDir := range inputDirs {
			dirPath := filepath.Join(dir, inputDir)
			if err := os.Mkdir(dirPath, 0o700); err != nil {
				t.Fatal(err)
			}
		}

		for _, inputFile := range inputFiles {
			filePath := filepath.Join(dir, inputFile)
			if err := os.WriteFile(filePath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}

		openDir := func(name string) (ReadDirCloser, error) {
			return os.Open(name)
		}
		skipDir := false
		var gotDirs, gotFiles []string
		for de, err := range NewReadDirIterRecursiveSorted(dir, openDir, &skipDir) {
			if err != nil {
				t.Fatal(err)
			}
			dePath := path.Join(de.Dir(), de.Entry().Name())
			if de.Entry().IsDir() {
				if dePath == filepath.Join(dir, "a", "b") {
					skipDir = true
					continue
				}
				gotDirs = append(gotDirs, dePath)
			} else {
				gotFiles = append(gotFiles, dePath)
			}
		}

		if !slices.Equal(gotDirs, wantDirs) {
			t.Errorf("dirs mismatch,\n got=%v,\nwant=%v", gotDirs, wantDirs)
		}
		if !slices.Equal(gotFiles, wantFiles) {
			t.Errorf("files mismatch,\n got=%v,\nwant=%v", gotFiles, wantFiles)
		}
	})
	t.Run("breakLoop", func(t *testing.T) {
		dir := tempDir(t)

		inputDirs := []string{
			"a",
			filepath.Join("a", "b"),
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "d"),
		}
		inputFiles := []string{
			filepath.Join("a", "b", "c", "f2"),
			filepath.Join("a", "f1"),
			filepath.Join("a", "f2"),
		}

		wantDirs := []string{
			filepath.Join(dir, "a"),
			filepath.Join(dir, "a", "b"),
			filepath.Join(dir, "a", "b", "c"),
		}
		wantFiles := []string{}

		for _, inputDir := range inputDirs {
			dirPath := filepath.Join(dir, inputDir)
			if err := os.Mkdir(dirPath, 0o700); err != nil {
				t.Fatal(err)
			}
		}

		for _, inputFile := range inputFiles {
			filePath := filepath.Join(dir, inputFile)
			if err := os.WriteFile(filePath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}

		openDir := func(name string) (ReadDirCloser, error) {
			return os.Open(name)
		}
		var gotDirs, gotFiles []string
		i := 0
		for de, err := range NewReadDirIterRecursiveSorted(dir, openDir, nil) {
			if err != nil {
				t.Fatal(err)
			}
			dePath := path.Join(de.Dir(), de.Entry().Name())
			if de.Entry().IsDir() {
				gotDirs = append(gotDirs, dePath)
			} else {
				gotFiles = append(gotFiles, dePath)
			}
			i++
			if i == 3 {
				break
			}
		}

		if !slices.Equal(gotDirs, wantDirs) {
			t.Errorf("dirs mismatch,\n got=%v,\nwant=%v", gotDirs, wantDirs)
		}
		if !slices.Equal(gotFiles, wantFiles) {
			t.Errorf("files mismatch,\n got=%v,\nwant=%v", gotFiles, wantFiles)
		}
	})
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
