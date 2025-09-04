package orto

import "github.com/anknetau/orto/git"

//go:generate go run golang.org/x/tools/cmd/stringer -type=ChangeKind
type ChangeKind int

const (
	ChangeKindAdded ChangeKind = iota
	ChangeKindDeleted
	ChangeKindUnchanged
	ChangeKindModified
	ChangeKindIgnoredByGit
	ChangeKindIgnoredByOrto
)

type Change struct {
	Kind   ChangeKind
	FsFile *FSFile
	Blob   *git.Blob
}
