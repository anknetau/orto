package orto

type Change interface {
	Kind() ChangeKind
}

type ChangeKind int

const (
	AddedKind ChangeKind = iota
	DeletedKind
	UnchangedKind
	ModifiedKind
	IgnoredByGitKind
	IgnoredByOrtoKind
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=ChangeKind

func (Added) Kind() ChangeKind         { return AddedKind }
func (Deleted) Kind() ChangeKind       { return DeletedKind }
func (Unchanged) Kind() ChangeKind     { return UnchangedKind }
func (Modified) Kind() ChangeKind      { return ModifiedKind }
func (IgnoredByGit) Kind() ChangeKind  { return IgnoredByGitKind }
func (IgnoredByOrto) Kind() ChangeKind { return IgnoredByOrtoKind }

type Added struct {
	FsFile FSFile
}

type Deleted struct {
	GitFile GitFile
}

type Unchanged struct {
	FsFile  FSFile
	GitFile GitFile
}

type Modified struct {
	FsFile  FSFile
	GitFile GitFile
}
type IgnoredByGit struct {
	FsFile FSFile
}

type IgnoredByOrto struct {
	FsFile FSFile
}
