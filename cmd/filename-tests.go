package main

import (
	"log"
	"os"
	"path/filepath"
)

func tryCreateUnlinked(dir, name string) error {
	p := filepath.Join(dir, name)

	// O_EXCL ensures no accidental overwrite if a file appears between checks.
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err // EINVAL/ENAMETOOLONG/EPERM, or EEXIST if collision
	}
	// Immediately remove the directory entry; file persists only as an open fd.
	_ = os.Remove(p)
	_ = f.Close()
	return nil
}

func main() {
	err := tryCreateUnlinked("./", "a")
	//open, err := os.Open("aá/a")
	if err != nil {
		log.Fatal(err)
	}
	//defer open.Close()
}
