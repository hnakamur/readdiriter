package readdiriter

import (
	"log"
	"os"
)

func ExampleNewReadDirIter() {
	file, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// See https://pkg.go.dev/os@latest#File.ReadDir for n
	const n = 0
	for entry, err := range NewReadDirIter(file, n) {
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("entry=%+v", entry)
	}
}
