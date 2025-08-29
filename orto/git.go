package orto

import (
	"log"
	"path/filepath"

	"github.com/anknetau/orto/fp"
)

// GitFile is a file reference in the git world.
type GitFile struct {
	CleanPath string
	path      string
	checksum  string
	mode      GitMode
}

type GitMode string

const (
	Directory  GitMode = "40000"
	File       GitMode = "100644"
	Executable GitMode = "100755"
	Symlink    GitMode = "120000"
	Submodule  GitMode = "160000"
)

func IsValidGitMode(mode string) bool {
	m := GitMode(mode)
	return m == Directory || m == File || m == Executable || m == Symlink || m == Submodule
}

func IsSupportedGitMode(mode string) bool {
	m := GitMode(mode)
	// TODO: will we ever need Directory? Probably not because we are looking just for files.
	return IsValidGitMode(mode) && m != Symlink && m != Submodule && m != Directory
}

func MakeGitFile(objectType string, path string, checksum string, mode string) GitFile {
	Filepath := filepath.Clean(path)
	if !IsValidGitMode(mode) {
		log.Fatal("Invalid git mode: " + mode)
	}
	if !IsSupportedGitMode(mode) {
		log.Fatal("Unsupported git mode: " + mode)
	}
	// When `git ls-tree` is passed -r, it will recurse and not show trees, but resolve the blobs within instead.
	if objectType != "blob" {
		log.Fatal("Submodules are not supported: " + objectType)
	}
	fp.ValidFilePathForOrtoOrDie(Filepath)
	gitFile := GitFile{Filepath, path, checksum, GitMode(mode)}
	if checksumGetAlgo(gitFile.checksum) == UNKNOWN {
		log.Fatal("Don't know how this was hashed: " + checksum)
	}
	return gitFile
}
