package git

import (
	"log"
	"path/filepath"

	"github.com/anknetau/orto/fp"
)

// Blob is a file reference in the git world.
type Blob struct {
	CleanPath string
	Path      string
	Checksum  fp.Checksum
	Mode      Mode
}

type Mode string

func NewMode(mode string) Mode {
	if !IsValidGitMode(mode) {
		log.Fatal("Invalid git mode: " + mode)
	}
	if !IsSupportedGitMode(mode) {
		log.Fatal("Unsupported git mode: " + mode)
	}
	return Mode(mode)
}

const (
	Deleted    Mode = "000000"
	Directory  Mode = "40000"
	File       Mode = "100644"
	Executable Mode = "100755"
	Symlink    Mode = "120000"
	Submodule  Mode = "160000"
)

func IsValidGitMode(mode string) bool {
	m := Mode(mode)
	return m == Directory || m == File || m == Executable || m == Symlink || m == Submodule || m == Deleted
}

func IsSupportedGitMode(mode string) bool {
	m := Mode(mode)
	// TODO: will we ever need Directory? Probably not because we are looking just for files.
	return m == File || m == Executable || m == Deleted
}

func NewGitFile(objectType string, path string, checksum fp.Checksum, mode Mode) Blob {
	Filepath := filepath.Clean(path)
	// When `git ls-tree` is passed -r, it will recurse and not show trees, but resolve the blobs within instead.
	if objectType != "blob" {
		log.Fatal("Submodules are not supported: " + objectType)
	}
	fp.ValidFilePathForOrtoOrDie(Filepath)
	return Blob{Filepath, path, checksum, mode}
}
