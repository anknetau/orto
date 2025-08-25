package main

import (
	"log"
	"os"
	"path/filepath"
)

type FSFile struct {
	filepath string
	root     string
	entry    os.DirEntry
}

func fsReadDir(root string) []FSFile {
	var entries []FSFile
	var err = filepath.WalkDir(root, func(root string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fsFile := FSFile{
			filepath: filepath.Clean(root),
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

func fsIndexByName(fsFiles []FSFile) map[string]FSFile {
	fsFileIndex := make(map[string]FSFile, len(fsFiles))
	for _, fsFile := range fsFiles {
		fsFileIndex[fsFile.filepath] = fsFile
	}
	return fsFileIndex
}
