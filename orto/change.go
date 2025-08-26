package orto

type Change interface {
	Type() ChangeType
}

type ChangeType int

const (
	AddedType ChangeType = iota
	DeletedType
	UnchangedType
	ModifiedType
	IgnoredByGitType
	IgnoredByOrtoType
)

func (Added) Type() ChangeType         { return AddedType }
func (Deleted) Type() ChangeType       { return DeletedType }
func (Unchanged) Type() ChangeType     { return UnchangedType }
func (Modified) Type() ChangeType      { return ModifiedType }
func (IgnoredByGit) Type() ChangeType  { return IgnoredByGitType }
func (IgnoredByOrto) Type() ChangeType { return IgnoredByOrtoType }

type Added struct {
	FsFile *FSFile
}

type Deleted struct {
	GitFile *GitFile
}

type Unchanged struct {
	FsFile  *FSFile
	GitFile *GitFile
}

type Modified struct {
	FsFile  *FSFile
	GitFile *GitFile
}
type IgnoredByGit struct {
	FsFile *FSFile
}

type IgnoredByOrto struct {
	FsFile *FSFile
}
