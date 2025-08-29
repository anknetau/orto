package orto

import (
	"os"
	"path/filepath"

	"github.com/anknetau/orto/fp"
)

// FSFile is an actual file system file.
type FSFile struct {
	CleanPath string
	path      string
	dirEntry  os.DirEntry
}

func FsReadDir(root string) []FSFile {
	var entries []FSFile
	var err = filepath.WalkDir(root, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dirEntry.IsDir() {
			return nil
		}
		fp.ValidFilePathForOrtoOrDie(path)
		Filepath := filepath.Clean(path)
		fsFile := FSFile{Filepath, path, dirEntry}
		entries = append(entries, fsFile)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return entries
}
