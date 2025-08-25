package main

type GitFile struct {
	filepath string
	path     string
	checksum string
	mode     string
}

func gitIndexByName(gitFiles []GitFile) map[string]GitFile {
	gitFileIndex := make(map[string]GitFile, len(gitFiles))
	for _, gitFile := range gitFiles {
		gitFileIndex[gitFile.filepath] = gitFile
	}
	return gitFileIndex
}
