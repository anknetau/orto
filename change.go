package main

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
	fsFile *FSFile
}

type Deleted struct {
	gitFile *GitFile
}

type Unchanged struct {
	fsFile  *FSFile
	gitFile *GitFile
}

type Modified struct {
	fsFile  *FSFile
	gitFile *GitFile
}
type IgnoredByGit struct {
	fsFile *FSFile
}

type IgnoredByOrto struct {
	fsFile *FSFile
}
