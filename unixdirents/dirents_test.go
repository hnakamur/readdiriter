package unixdirents

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

func TestDirents(t *testing.T) {
	testCases := []int{
		0,
		1,
		10,
		100,
		1_000,
		10_000,
		100_000,
	}
	for _, tt := range testCases {
		t.Run(fmt.Sprintf("n=%d", tt), func(t *testing.T) {
			dir := t.TempDir()
			const filenamePrefix = "file"
			want := make([]string, 0, tt)
			for i := range tt {
				filename := fmt.Sprintf("%s%d", filenamePrefix, i)
				filepath := filepath.Join(dir, filename)
				if err := os.WriteFile(filepath, []byte{}, 0o600); err != nil {
					t.Fatal(err)
				}
				want = append(want, filename)
			}

			f, err := os.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			got := make([]string, 0, tt)
			var buf [4096]byte
			for de, err := range Dirents(int(f.Fd()), buf[:]) {
				if err != nil {
					t.Fatal(err)
				}

				filename := string(de.Name)
				if !strings.HasPrefix(filename, filenamePrefix) {
					t.Fatalf("filename %s should have prefix: %s",
						filename, filenamePrefix)
				}
				got = append(got, filename)

				if gotType, wantType := de.Type, DTypeRegularFile; gotType != wantType {
					t.Errorf("directory entry type mismatch for %s, got=%d, want=%d",
						filename, gotType, wantType)
				}
			}

			getNumPart := func(filename string) int {
				i, err := strconv.Atoi(strings.TrimPrefix(filename, filenamePrefix))
				if err != nil {
					t.Fatalf("filename %s must have integer suffix: %s", filename, err)
				}
				return i
			}
			slices.SortFunc(got, func(a, b string) int {
				return cmp.Compare(getNumPart(a), getNumPart(b))
			})

			if diff := gocmp.Diff(want, got); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
