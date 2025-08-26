package main

import (
	"log"
	"os"
	"strings"

	"github.com/anknetau/orto/orto"
)

func main() {
	destination := "/Users/ank/dev/orto_dest_test"
	destinationAsBaseForCopying := destination + "/" // TODO
	orto.CheckDestination(destination)

	//err := os.Chdir("/Users/ank/dev/accounting/accounting")
	//err := os.Chdir("/Users/ank/dev/mirrors")
	err := os.Chdir("/Users/ank/dev/orto")
	if err != nil {
		log.Fatal(err)
	}
	// FS Files
	fsFiles := orto.FsReadDir("./") // TODO "./"
	gitFiles := orto.GitRunGetTreeForHead()
	gitStatus := orto.GitRunStatus()

	fsFileIndex := orto.Index(fsFiles, func(file orto.FSFile) string {
		return file.Filepath
	})
	gitFileIndex := orto.Index(gitFiles, func(file orto.GitFile) string {
		return file.Filepath
	})
	var gitIgnoredFiles []string
	for _, status := range gitStatus {
		switch v := status.(type) {
		case *orto.IgnoredStatusLine:
			gitIgnoredFiles = append(gitIgnoredFiles, v.Path)
		default:
		}
	}
	gitIgnoredFilesIndex := orto.Index(gitIgnoredFiles, func(file string) string {
		return file
	})

	common, left, right := orto.CompareFiles(fsFiles, gitFiles, fsFileIndex, gitFileIndex)

	// Filter out ignored
	//var aLeft []FSFile
	//for _, f := range left {
	//	if _, found := gitIgnoredFilesIndex[f.filepath]; !found {
	//		aLeft = append(aLeft, f)
	//	}
	//}

	var allChanges []orto.Change
	for _, f := range left {
		change := orto.ComparePair(nil, &f, gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}
	for _, f := range right {
		change := orto.ComparePair(&f, nil, gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}
	for _, c := range common {
		change := orto.ComparePair(&c.GitFile, &c.FsFile, gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}

	for _, c := range allChanges {
		switch c := c.(type) {
		case orto.Added:
			orto.CopyFile(c.FsFile.Filepath, destinationAsBaseForCopying+c.FsFile.Filepath) // TODO: /
			orto.ChangePrint(c)
		case orto.Modified:
			orto.CopyFile(c.FsFile.Filepath, destinationAsBaseForCopying+c.FsFile.Filepath) // TODO: /
			orto.ChangePrint(c)
		case orto.Deleted:
			orto.ChangePrint(c)
		}
	}
	println("----")
	for _, c := range allChanges {
		if c.Type() == orto.UnchangedType {
			orto.ChangePrint(c)
		}
	}
	for _, c := range allChanges {
		if c.Type() == orto.IgnoredByGitType {
			orto.ChangePrint(c)
		}
	}
	allDotGitChanges := true
	for _, c := range allChanges {
		switch change := c.(type) {
		case orto.IgnoredByOrto:
			// TODO: fix this:
			if !strings.HasPrefix(change.FsFile.Filepath, ".git/") {
				allDotGitChanges = false
				break
			}
		}
	}
	if allDotGitChanges {
		println("⛔︎ OrtoIgnored", ".git/**")
	} else {
		for _, c := range allChanges {
			switch c.(type) {
			case orto.IgnoredByOrto:
				orto.ChangePrint(c)
			}
		}
	}

	//println("---")
	//for _, fsFile := range fsFiles {
	//	gitFile, ok := gitFileIndex[fsFile.filepath]
	//	pGitFile := &gitFile
	//	if !ok {
	//		pGitFile = nil
	//	}
	//	compare(pGitFile, &fsFile)
	//}
	//println("---")
	//for _, gitFile := range gitFiles {
	//	fsFile, ok := fsFileIndex[gitFile.filepath]
	//	pFsFile := &fsFile
	//	if !ok {
	//		pFsFile = nil
	//	}
	//	compare(&gitFile, pFsFile)
	//}
}
