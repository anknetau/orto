package orto

import (
	"os"
	"path/filepath"

	"github.com/anknetau/orto/fp"
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

		fp.ValidFilePathForOrtoOrDie(root)

		fsFile := FSFile{
			Filepath: filepath.Clean(root),
			root:     root,
			entry:    info,
		}
		entries = append(entries, fsFile)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return entries
}
