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
	Checksum  string
	Mode      Mode
}

type Mode string

const (
	Directory  Mode = "40000"
	File       Mode = "100644"
	Executable Mode = "100755"
	Symlink    Mode = "120000"
	Submodule  Mode = "160000"
)

func IsValidGitMode(mode string) bool {
	m := Mode(mode)
	return m == Directory || m == File || m == Executable || m == Symlink || m == Submodule
}

func IsSupportedGitMode(mode string) bool {
	m := Mode(mode)
	// TODO: will we ever need Directory? Probably not because we are looking just for files.
	return IsValidGitMode(mode) && m != Symlink && m != Submodule && m != Directory
}

func NewGitFile(objectType, path, checksum, mode string) Blob {
	Filepath := filepath.Clean(path)
	if !IsValidGitMode(mode) {
		log.Fatal("Invalid git mode: " + mode)
	}
	if !IsSupportedGitMode(mode) {
		log.Fatal("Unsupported git mode: " + mode + " for " + path)
	}
	// When `git ls-tree` is passed -r, it will recurse and not show trees, but resolve the blobs within instead.
	if objectType != "blob" {
		log.Fatal("Submodules are not supported: " + objectType)
	}
	fp.ValidFilePathForOrtoOrDie(Filepath)
	gitFile := Blob{Filepath, path, checksum, Mode(mode)}
	if fp.ChecksumGetAlgo(gitFile.Checksum) == fp.UNKNOWN {
		log.Fatal("Don't know how this was hashed: " + checksum)
	}
	return gitFile
}
