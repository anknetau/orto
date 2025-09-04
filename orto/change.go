package orto

import "github.com/anknetau/orto/git"

//go:generate go run golang.org/x/tools/cmd/stringer -type=ChangeKind
type ChangeKind int

const (
	AddedKind ChangeKind = iota
	DeletedKind
	UnchangedKind
	ModifiedKind
	IgnoredByGitKind
	IgnoredByOrtoKind
)

type Change struct {
	Kind   ChangeKind
	FsFile *FSFile
	Blob   *git.Blob
}
