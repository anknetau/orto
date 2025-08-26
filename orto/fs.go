package orto

import (
	"log"
	"os"
	"path/filepath"
)

type FSFile struct {
	Filepath string
	root     string
	entry    os.DirEntry
}

func FsReadDir(root string) []FSFile {
	var entries []FSFile
	var err = filepath.WalkDir(root, func(root string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fsFile := FSFile{
			Filepath: filepath.Clean(root),
			root:     root,
			entry:    info,
		}
		entries = append(entries, fsFile)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return entries
}
