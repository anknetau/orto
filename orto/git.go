package orto

// GitFile is a file reference in the git world.
type GitFile struct {
	CleanPath string
	path      string
	checksum  string
	mode      string
}
