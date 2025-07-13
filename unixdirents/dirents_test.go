package unixdirents

import (
	"cmp"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"golang.org/x/sys/unix"
)

func TestDirents(t *testing.T) {
	type myDirent struct {
		Ino  uint64
		Type DType
		Name string
	}

	cmpMyDirent := func(a, b myDirent) int { return cmp.Compare(a.Name, b.Name) }
	sortDirents := func(dents []myDirent) { slices.SortFunc(dents, cmpMyDirent) }

	allDirents := func(t *testing.T, dir string) []myDirent {
		f, err := os.Open(dir)
		if err != nil {
			t.Fatalf("open dir: %s, %s", dir, err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				t.Fatalf("close dir: %s, %s", dir, err)
			}
		}()

		var dents []myDirent
		var buf [4096]byte
		for de, err := range Dirents(int(f.Fd()), buf[:]) {
			if err != nil {
				t.Fatalf("error in Dirents: %s", err)
			}

			dents = append(dents, myDirent{
				Ino:  de.Ino,
				Type: de.Type,
				Name: string(de.Name),
			})
		}
		return dents
	}

	getInodeFromFileInfo := func(fi fs.FileInfo) uint64 {
		if fiSys, ok := fi.Sys().(*syscall.Stat_t); ok {
			return fiSys.Ino
		} else {
			panic("file info sys is not *syscall.Stat_t")
		}
	}

	createRegularFile := func(t *testing.T, dir string) myDirent {
		file, err := os.CreateTemp(dir, "file")
		if err != nil {
			t.Fatalf("create regular file: %s", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				t.Fatalf("close regular file: %s", err)
			}
		}()

		fi, err := file.Stat()
		if err != nil {
			t.Fatalf("stat file: %s", err)
		}
		return myDirent{
			Ino:  getInodeFromFileInfo(fi),
			Type: DTypeRegularFile,
			Name: filepath.Base(file.Name()),
		}
	}

	createRegularFiles := func(t *testing.T, dir string, n int) []myDirent {
		var dents []myDirent
		for range n {
			dent := createRegularFile(t, dir)
			dents = append(dents, dent)
		}
		return dents
	}

	tempFilepath := func(t *testing.T, dir, pattern string) string {
		file, err := os.CreateTemp(dir, pattern)
		if err != nil {
			t.Fatalf("create temporary file: %s", err)
		}
		if err := file.Close(); err != nil {
			t.Fatalf("close temporary file: %s", err)
		}
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("remove temporary file: %s", err)
		}
		return file.Name()
	}

	createNamedPipe := func(t *testing.T, dir string) myDirent {
		path := tempFilepath(t, dir, "namedpipe")
		if err := unix.Mkfifo(path, 0o600); err != nil {
			t.Fatalf("create named pipe: %s", err)
		}
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat fifo: %s", err)
		}
		return myDirent{
			Ino:  getInodeFromFileInfo(fi),
			Type: DTypeNamedPipe,
			Name: filepath.Base(path),
		}
	}

	createSymlink := func(t *testing.T, dir string) myDirent {
		linkPath := tempFilepath(t, dir, "symlink")
		targetPath := tempFilepath(t, dir, "symlinktarget")
		if err := os.Symlink(targetPath, linkPath); err != nil {
			t.Fatalf("create named pipe: %s", err)
		}
		fi, err := os.Lstat(linkPath)
		if err != nil {
			t.Fatalf("stat fifo: %s", err)
		}
		return myDirent{
			Ino:  getInodeFromFileInfo(fi),
			Type: DTypeSymlink,
			Name: filepath.Base(linkPath),
		}
	}

	createSocket := func(t *testing.T, dir string) myDirent {
		path := tempFilepath(t, dir, "socket")
		listener, err := net.Listen("unix", path)
		if err != nil {
			t.Fatalf("create socket file: %s", err)
		}
		t.Cleanup(func() {
			if err := listener.Close(); err != nil {
				t.Fatalf("close socket file: %s", err)
			}
		})
		fi, err := os.Lstat(path)
		if err != nil {
			t.Fatalf("stat fifo: %s", err)
		}
		return myDirent{
			Ino:  getInodeFromFileInfo(fi),
			Type: DTypeSocket,
			Name: filepath.Base(path),
		}
	}

	createDirectory := func(t *testing.T, dir string) myDirent {
		path, err := os.MkdirTemp(dir, "dir")
		if err != nil {
			t.Fatalf("create directory: %s", err)
		}
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat fifo: %s", err)
		}
		return myDirent{
			Ino:  getInodeFromFileInfo(fi),
			Type: DTypeDir,
			Name: filepath.Base(path),
		}
	}

	testCases := []int{
		0,
		1,
		10,
		100,
		1_000,
		10_000,
		100_000,
	}
	for _, fileCount := range testCases {
		t.Run(fmt.Sprintf("file_n=%d", fileCount), func(t *testing.T) {
			dir := t.TempDir()

			want := createRegularFiles(t, dir, fileCount)
			got := allDirents(t, dir)
			sortDirents(got)
			sortDirents(want)
			if diff := gocmp.Diff(want, got); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("variousDTypes", func(t *testing.T) {
		dir := t.TempDir()
		want := []myDirent{
			createRegularFile(t, dir),
			createSymlink(t, dir),
			createNamedPipe(t, dir),
			createSocket(t, dir),
			createDirectory(t, dir),
		}
		got := allDirents(t, dir)
		sortDirents(got)
		sortDirents(want)
		if diff := gocmp.Diff(want, got); diff != "" {
			t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
		}
	})
}
