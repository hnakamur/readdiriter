package unixdirents

import (
	"log"
	"os"
)

func ExampleDirents() {
	f, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var buf [4096]byte
	for de, err := range Dirents(int(f.Fd()), buf[:]) {
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("directory entry=%+v", de)
	}
}
