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

func NewFSFile(path string, dirEntry os.DirEntry) FSFile {
	cleanPath := filepath.Clean(path)
	fp.ValidFilePathForOrtoOrDie(cleanPath)
	return FSFile{cleanPath, path, dirEntry}
}

func FsReadDir(root string) []FSFile {
	var entries []FSFile
	var err = filepath.WalkDir(root, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if !filepath.IsLocal(relPath) {
			panic(relPath)
		}
		if dirEntry.IsDir() {
			// TODO: this ignores .git
			if dirEntry.Name() == ".git" && relPath == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		// TODO: add filepath.IsLocal() where needed, for security
		fsFile := NewFSFile(relPath, dirEntry)
		entries = append(entries, fsFile)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return entries
}
