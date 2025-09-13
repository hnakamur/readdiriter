// Package unixdirents provides functions which returns a iterator
// over directory entries in the specified directory on Unix platforms.
package unixdirents

import (
	"bytes"
	"encoding/binary"
	"iter"
	"unsafe"

	"golang.org/x/sys/unix"
)

// DType represents the type of a directory entry.
type DType uint8

const (
	// DTypeUnknown indicates an unknown file type.
	DTypeUnknown DType = 0
	// DTypeNamedPipe represents a named pipe (FIFO).
	DTypeNamedPipe DType = 1
	// DTypeCharDevice represents a character device.
	DTypeCharDevice DType = 2
	// DTypeDir represents a directory.
	DTypeDir DType = 4
	// DTypeBlockDevice represents a block device.
	DTypeBlockDevice DType = 6
	// DTypeRegularFile represents a regular file.
	DTypeRegularFile DType = 8
	// DTypeSymlink represents a symbolic link.
	DTypeSymlink DType = 10
	// DTypeSocket represents a UNIX domain socket.
	DTypeSocket DType = 12
)

// DirentInfo is a directory entry info.
type DirentInfo struct {
	// Ino is the 64-bit inode nubmer.
	Ino uint64

	// Type is the file type.
	Type DType

	// Name is the filename.
	//
	// The slice is only valid until the next modification of the underlying
	// buffer, which typically occurs on the next call to unix.ReadDirent
	// (e.g., when iterating with the Dirents method).
	Name []byte
}

// Dirents returns an iterator over directory entries within the specified file
// descriptor fd.
// The provided buffer buf is used for reading directory data.
//
// Note: The directory entries are not in lexical order.
func Dirents(fd int, buf []byte) iter.Seq2[DirentInfo, error] {
	return func(yield func(DirentInfo, error) bool) {
		for {
			n, err := unix.ReadDirent(fd, buf)
			if err != nil {
				yield(DirentInfo{}, err)
				return
			}
			if n == 0 {
				return
			}

			for de := range direntsInBuf(buf[:n]) {
				if !yield(de, nil) {
					return
				}
			}
		}
	}
}

// direntsInBuf returns an iterator over directory entries found within the
// given buffer.
// This function is an internal helper.
func direntsInBuf(buf []byte) iter.Seq[DirentInfo] {
	return func(yield func(DirentInfo) bool) {
		// This code is based on
		// https://cs.opensource.google/go/x/sys/+/refs/tags/v0.34.0:unix/dirent.go;l=64
		// https://cs.opensource.google/go/x/sys/+/refs/tags/v0.34.0:unix/ztypes_linux_amd64.go;l=102-109
		const (
			inoOff    = int(unsafe.Offsetof(unix.Dirent{}.Ino))
			reclenOff = int(unsafe.Offsetof(unix.Dirent{}.Reclen))
			typeOff   = int(unsafe.Offsetof(unix.Dirent{}.Type))
			nameOff   = int(unsafe.Offsetof(unix.Dirent{}.Name))

			inoSize    = int(unsafe.Sizeof(unix.Dirent{}.Ino))
			reclenSize = int(unsafe.Sizeof(unix.Dirent{}.Reclen))
			typeSize   = int(unsafe.Sizeof(unix.Dirent{}.Type))
		)

		if inoSize != 8 {
			panic("unsupported unix.Dirent{} size")
		}
		if reclenSize != 2 {
			panic("unsupported unix.Reclen{} size")
		}
		if typeSize != 1 {
			panic("unsupported unix.Type{} size")
		}

		for len(buf) >= reclenOff+reclenSize {
			reclen := binary.NativeEndian.Uint16(buf[reclenOff:])
			if reclen == 0 || int(reclen) > len(buf) {
				return
			}
			rec := buf[:reclen]
			buf = buf[reclen:]

			ino := binary.NativeEndian.Uint64(rec[inoOff:])
			if ino == 0 { // File absent in directory.
				continue
			}
			dType := rec[typeOff]
			name := rec[nameOff:]
			if i := bytes.IndexByte(name, byte('\x00')); i != -1 {
				name = name[:i]
			}

			// Check for useless names before allocating a string.
			if bytes.Equal(name, []byte(".")) || bytes.Equal(name, []byte("..")) {
				continue
			}

			di := DirentInfo{
				Ino:  ino,
				Type: DType(dType),
				Name: name,
			}
			if !yield(di) {
				return
			}
		}
	}
}
