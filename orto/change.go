package orto

import "github.com/anknetau/orto/git"

type ChangeKind int

const (
	AddedKind ChangeKind = iota
	DeletedKind
	UnchangedKind
	ModifiedKind
	IgnoredByGitKind
	IgnoredByOrtoKind
)

//go:generate stringer -type=ChangeKind

type Change struct {
	Kind   ChangeKind
	FsFile *FSFile
	Blob   *git.Blob
}
