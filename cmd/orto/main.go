package main

import (
	"github.com/anknetau/orto/orto"
	"log"
	"os"
	"strings"
)

func main() {
	destination := "/Users/ank/dev/orto_dest_test"
	destinationAsBaseForCopying := destination + "/" // TODO
	CheckDestination(destination)

	//err := os.Chdir("/Users/ank/dev/accounting/accounting")
	//err := os.Chdir("/Users/ank/dev/mirrors")
	err := os.Chdir("/Users/ank/dev/orto")
	if err != nil {
		log.Fatal(err)
	}
	// FS Files
	fsFiles := fsReadDir("./") // TODO "./"
	gitFiles := gitRunGetTreeForHead()
	gitStatus := gitRunStatus()

	fsFileIndex := index(fsFiles, func(file FSFile) string {
		return file.filepath
	})
	gitFileIndex := index(gitFiles, func(file GitFile) string {
		return file.filepath
	})
	var gitIgnoredFiles []string
	for _, status := range gitStatus {
		switch v := status.(type) {
		case *IgnoredStatusLine:
			gitIgnoredFiles = append(gitIgnoredFiles, v.path)
		default:
		}
	}
	gitIgnoredFilesIndex := index(gitIgnoredFiles, func(file string) string {
		return file
	})

	common, left, right := compareFiles(fsFiles, gitFiles, fsFileIndex, gitFileIndex)

	// Filter out ignored
	//var aLeft []FSFile
	//for _, f := range left {
	//	if _, found := gitIgnoredFilesIndex[f.filepath]; !found {
	//		aLeft = append(aLeft, f)
	//	}
	//}

	var allChanges []Change
	for _, f := range left {
		change := comparePair(nil, &f, gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}
	for _, f := range right {
		change := comparePair(&f, nil, gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}
	for _, c := range common {
		change := comparePair(&c.gitFile, &c.fsFile, gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}

	for _, c := range allChanges {
		switch c := c.(type) {
		case Added:
			copyFile(c.fsFile.filepath, destinationAsBaseForCopying+c.fsFile.filepath) // TODO: /
			changePrint(c)
		case Modified:
			copyFile(c.fsFile.filepath, destinationAsBaseForCopying+c.fsFile.filepath) // TODO: /
			changePrint(c)
		case Deleted:
			changePrint(c)
		}
	}
	println("----")
	for _, c := range allChanges {
		if c.Type() == UnchangedType {
			changePrint(c)
		}
	}
	for _, c := range allChanges {
		if c.Type() == IgnoredByGitType {
			changePrint(c)
		}
	}
	allDotGitChanges := true
	for _, c := range allChanges {
		switch change := c.(type) {
		case IgnoredByOrto:
			// TODO: fix this:
			if !strings.HasPrefix(change.fsFile.filepath, ".git/") {
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
			case IgnoredByOrto:
				changePrint(c)
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
