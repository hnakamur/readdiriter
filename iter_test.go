package readdiriter

import (
	"errors"
	"io"
	"io/fs"
	"testing"
)

//
// Mock Implementations
//

// mockDirEntry is a mock implementation that satisfies the fs.DirEntry interface.
type mockDirEntry struct {
	name    string
	isDir   bool
	fileSys fs.FileInfo
}

// Name returns the name of the directory entry.
func (m mockDirEntry) Name() string {
	return m.name
}

// IsDir returns true if the entry is a directory.
func (m mockDirEntry) IsDir() bool {
	return m.isDir
}

// Type returns the file mode of the entry.
func (m mockDirEntry) Type() fs.FileMode {
	if m.isDir {
		return fs.ModeDir
	}
	return 0
}

// Info returns the FileInfo for the entry.
func (m mockDirEntry) Info() (fs.FileInfo, error) {
	return m.fileSys, nil
}

// mockFileInfo is a mock implementation that satisfies the fs.FileInfo interface.
type mockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime string // Simplistic for testing, could be time.Time
	isDir   bool
	sys     any
}

// Name returns the base name of the file.
func (m mockFileInfo) Name() string { return m.name }

// Size returns the size of the file in bytes.
func (m mockFileInfo) Size() int64 { return m.size }

// Mode returns the file mode bits.
func (m mockFileInfo) Mode() fs.FileMode { return m.mode }

// ModTime returns the modification time of the file. (Not fully implemented for test)
func (m mockFileInfo) ModTime() {}

// IsDir returns true if the file is a directory.
func (m mockFileInfo) IsDir() bool { return m.isDir }

// Sys returns the underlying data source (can be nil).
func (m mockFileInfo) Sys() any { return m.sys }

// mockReadDirFile is a mock implementation that satisfies the fs.ReadDirFile interface.
type mockReadDirFile struct {
	entries   []fs.DirEntry
	readCount int
	errOnCall error
	eofAfterN int // Returns io.EOF after n calls to ReadDir
}

// Read is not relevant for this test, returns 0 and nil.
func (m *mockReadDirFile) Read(_ []byte) (int, error) {
	return 0, nil
}

// Close is not relevant for this test, returns nil.
func (m *mockReadDirFile) Close() error {
	return nil
}

// Stat is not relevant for this test, returns nil and nil.
func (m *mockReadDirFile) Stat() (fs.FileInfo, error) {
	return nil, nil
}

// ReadDir simulates reading directory entries.
// It returns a slice of DirEntry and an error.
// The behavior changes based on 'n' and internal state (readCount, errOnCall, eofAfterN).
func (m *mockReadDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	m.readCount++
	// Simulate an error on the first call if errOnCall is set
	if m.errOnCall != nil && m.readCount == 1 {
		return nil, m.errOnCall
	}
	// Simulate EOF after a specified number of calls
	if m.eofAfterN > 0 && m.readCount > m.eofAfterN {
		return nil, io.EOF
	}

	if n <= 0 {
		// If n <= 0, return all entries at once
		if m.readCount > 1 { // Return EOF on subsequent calls if n <= 0
			return nil, io.EOF
		}
		return m.entries, nil
	}

	// For n > 0, return entries in chunks
	start := (m.readCount - 1) * n
	if start >= len(m.entries) {
		return nil, io.EOF
	}

	end := start + n
	if end > len(m.entries) {
		end = len(m.entries)
	}
	return m.entries[start:end], nil
}

//
// Test Cases
//

// TestNewReadDirIter tests the NewReadDirIter function with various scenarios.
func TestNewReadDirIter(t *testing.T) {
	tests := []struct {
		name          string
		mockFile      *mockReadDirFile
		n             int
		breakAfter    int
		expectedNames []string
		expectedErr   error
	}{
		{
			name: "n=0_read_all_files_at_once",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "file1"},
					mockDirEntry{name: "file2"},
					mockDirEntry{name: "dir1", isDir: true},
				},
			},
			n:             0,
			breakAfter:    -1,
			expectedNames: []string{"file1", "file2", "dir1"},
			expectedErr:   nil,
		},
		{
			name: "n=0_no_files_exist",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{},
			},
			n:             0,
			breakAfter:    -1,
			expectedNames: []string{},
			expectedErr:   nil,
		},
		{
			name: "n=0_break",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "file1"},
					mockDirEntry{name: "file2"},
					mockDirEntry{name: "dir1", isDir: true},
				},
			},
			n:             0,
			breakAfter:    1,
			expectedNames: []string{"file1", "file2"},
			expectedErr:   nil,
		},
		{
			name: "n=0_ReadDir_returns_error",
			mockFile: &mockReadDirFile{
				errOnCall: errors.New("permission denied"),
			},
			n:             0,
			breakAfter:    -1,
			expectedNames: []string{},
			expectedErr:   errors.New("permission denied"),
		},
		{
			name: "n>0_read_a_few_files",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "a.txt"},
					mockDirEntry{name: "b.txt"},
					mockDirEntry{name: "c.txt"},
					mockDirEntry{name: "d.txt"},
					mockDirEntry{name: "e.txt"},
				},
			},
			n:             2,
			breakAfter:    -1,
			expectedNames: []string{"a.txt", "b.txt", "c.txt", "d.txt", "e.txt"},
			expectedErr:   nil,
		},
		{
			name: "n>0_break",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "a.txt"},
					mockDirEntry{name: "b.txt"},
					mockDirEntry{name: "c.txt"},
					mockDirEntry{name: "d.txt"},
					mockDirEntry{name: "e.txt"},
				},
			},
			n:             2,
			breakAfter:    2,
			expectedNames: []string{"a.txt", "b.txt", "c.txt"},
			expectedErr:   nil,
		},
		{
			name: "n>0_entries_are_multiple_of_n",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "1.txt"},
					mockDirEntry{name: "2.txt"},
					mockDirEntry{name: "3.txt"},
					mockDirEntry{name: "4.txt"},
				},
			},
			n:             2,
			breakAfter:    -1,
			expectedNames: []string{"1.txt", "2.txt", "3.txt", "4.txt"},
			expectedErr:   nil,
		},
		{
			name: "n>0_no_entries_exist",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{},
			},
			n:             5,
			breakAfter:    -1,
			expectedNames: []string{},
			expectedErr:   nil,
		},
		{
			name: "n>0_ReadDir_returns_error_midway",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "item1"},
					mockDirEntry{name: "item2"},
				},
				errOnCall: errors.New("disk full"), // Error on the first ReadDir call
			},
			n:             1,
			breakAfter:    -1,
			expectedNames: []string{}, // No entries should be returned if an error occurs
			expectedErr:   errors.New("disk full"),
		},
		{
			name: "n>0_ReadDir_returns_EOF",
			mockFile: &mockReadDirFile{
				entries: []fs.DirEntry{
					mockDirEntry{name: "end1"},
					mockDirEntry{name: "end2"},
				},
				eofAfterN: 2, // Return EOF after 2nd ReadDir call
			},
			n:             1,
			breakAfter:    -1,
			expectedNames: []string{"end1", "end2"},
			expectedErr:   nil,
		},
		{
			name: "n>0_ReadDir_returns_EOF_on_first_call",
			mockFile: &mockReadDirFile{
				entries:   []fs.DirEntry{},
				eofAfterN: 0, // EOF on the very first call
			},
			n:             1,
			breakAfter:    -1,
			expectedNames: []string{},
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := NewReadDirIter(tt.mockFile, tt.n)

			var actualNames []string
			var actualErr error
			// Iterate through the sequence returned by NewReadDirIter
			i := 0
			for entry, err := range iter {
				if err != nil {
					actualErr = err
					break // Stop iteration on the first error
				}
				actualNames = append(actualNames, entry.Name())
				if tt.breakAfter != -1 && i == tt.breakAfter {
					break
				}
				i++
			}

			// Compare errors
			if tt.expectedErr != nil {
				if actualErr == nil {
					t.Fatalf("Expected error: %v, got nil", tt.expectedErr)
				}
				// Use errors.Is for wrapped errors or check error message directly
				if !errors.Is(actualErr, tt.expectedErr) && actualErr.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tt.expectedErr, actualErr)
				}
			} else if actualErr != nil {
				t.Fatalf("Expected no error, got: %v", actualErr)
			}

			// Compare results (directory entry names)
			if len(actualNames) != len(tt.expectedNames) {
				t.Fatalf("Expected %d entries, got %d", len(tt.expectedNames), len(actualNames))
			}
			for i, name := range actualNames {
				if name != tt.expectedNames[i] {
					t.Errorf("Entry at index %d mismatch. Expected: %s, Got: %s", i, tt.expectedNames[i], name)
				}
			}
		})
	}
}
