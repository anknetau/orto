package git

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/fp"
)

// Submodule reference, always points to a directory and the checksum is the commit in the separate repo.
// Its mode is always "160000" and its object type is always "commit"; additionally, "160000" and "commit" only ever
// appear together and for a submodule.
type Submodule struct {
	DirCleanPath string
	DirPath      string // relative path to directory as returned by git
	Checksum     fp.Checksum
}

// Blob is a file reference in the git world.
type Blob struct {
	CleanPath string
	Path      string // relative path as returned by git
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
	ModeDeleted    Mode = "000000"
	ModeDirectory  Mode = "40000"
	ModeFile       Mode = "100644"
	ModeExecutable Mode = "100755"
	ModeSymlink    Mode = "120000"
	ModeSubmodule  Mode = "160000"
)

const (
	ObjectTypeBlob   = "blob"
	ObjectTypeTree   = "tree"
	ObjectTypeCommit = "commit"
)

func IsValidGitMode(mode string) bool {
	m := Mode(mode)
	return m == ModeDirectory || m == ModeFile || m == ModeExecutable ||
		m == ModeSymlink || m == ModeSubmodule || m == ModeDeleted
}

func IsSupportedGitMode(mode string) bool {
	m := Mode(mode)
	// TODO: will we ever need Directory? Probably not because we are looking just for files.
	return m == ModeFile || m == ModeExecutable || m == ModeDeleted || m == ModeSubmodule
}

// Returns either Blob or Submodule, but never both.
func parseGetTreeLine(line string) (*Blob, *Submodule) {
	fields := strings.Split(line, "|>")
	//fmt.Printf("fields: %#v\n", fields)
	if len(fields) != 4 || len(fields[1]) == 0 || len(fields[2]) == 0 || len(fields[3]) == 0 {
		log.Fatal("Invalid line from git: " + line)
	}
	objectType := fields[0]
	checksum := fp.NewChecksum(fields[1])
	path := fields[2]
	mode := NewMode(fields[3])
	if objectType == ObjectTypeCommit {
		newSubmodule := NewSubmodule(objectType, path, checksum, mode)
		return nil, &newSubmodule
	} else if objectType == ObjectTypeBlob {
		newBlob := NewBlob(objectType, path, checksum, mode)
		return &newBlob, nil
	} else {
		// When `git ls-tree` is passed -r, it will recurse and not show trees, but resolve the blobs within instead.
		log.Fatal("Unsupported git object type: " + objectType)
		return nil, nil
	}
}

func NewBlob(objectType string, path string, checksum fp.Checksum, mode Mode) Blob {
	if !filepath.IsLocal(path) {
		log.Fatal("Git path is absolute or incorrect: " + path)
	}
	CleanPath := filepath.Clean(path)
	if objectType != ObjectTypeBlob {
		log.Fatal("Git object type is incorrect: " + objectType)
	}
	fp.ValidFilePathForOrtoOrDie(CleanPath)
	return Blob{CleanPath, path, checksum, mode}
}

func NewSubmodule(objectType string, path string, checksum fp.Checksum, mode Mode) Submodule {
	if !filepath.IsLocal(path) {
		log.Fatal("Git path is absolute or incorrect: " + path)
	}
	if objectType != ObjectTypeCommit {
		log.Fatal("Git object type is incorrect: " + objectType)
	}
	if mode != ModeSubmodule {
		log.Fatal("Git mode for submodule is incorrect: " + objectType)
	}
	return Submodule{
		DirCleanPath: filepath.Clean(path),
		DirPath:      path,
		Checksum:     checksum,
	}
}
