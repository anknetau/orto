package main

type Change interface {
	Type() string
}

type ChangeType int

const (
	AddedType ChangeType = iota
	DeletedType
	UnchangedType
	ModifiedType
)

func (Added) Type() ChangeType     { return AddedType }
func (Deleted) Type() ChangeType   { return DeletedType }
func (Unchanged) Type() ChangeType { return UnchangedType }
func (Modified) Type() ChangeType  { return ModifiedType }

type Added struct {
	fsFile FSFile
}

type Deleted struct {
	gitFile GitFile
}

type Unchanged struct {
	fsFile  FSFile
	gitFile GitFile
}

type Modified struct {
	fsFile  FSFile
	gitFile GitFile
}
