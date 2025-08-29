package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/orto"
)

func main() {
	params := orto.OrtoParameters{
		//err := os.Chdir("/Users/ank/dev/accounting/accounting")
		//err := os.Chdir("/Users/ank/dev/mirrors")
		Source:      "/Users/ank/dev/orto",
		Destination: "/Users/ank/dev/orto/dest", // TODO: this is within the source!
	}
	orto.Start(params)
	orto.CheckSource(params.Source)
	orto.CheckDestination(params.Destination)

	err := os.Chdir(params.Source)
	if err != nil {
		log.Fatal(err)
	}
	// FS Files
	fsFiles := orto.FsReadDir("./") // TODO "./"
	gitFiles := orto.GitRunGetTreeForHead()
	gitStatus := orto.GitRunStatus()

	fsFileIndex := orto.Index(fsFiles, func(file orto.FSFile) string {
		return file.CleanPath
	})
	gitFileIndex := orto.Index(gitFiles, func(file orto.GitFile) string {
		return file.CleanPath
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
		change := orto.ComparePair(nil, &f, gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}
	for _, f := range right {
		change := orto.ComparePair(&f, nil, gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}
	for _, c := range common {
		change := orto.ComparePair(&c.GitFile, &c.FsFile, gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}

	for _, c := range allChanges {
		switch c.Kind {
		case orto.AddedKind, orto.ModifiedKind, orto.DeletedKind:
			orto.PrintChange(c)
		}
	}
	println("----")
	for _, c := range allChanges {
		if c.Kind == orto.UnchangedKind {
			orto.PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == orto.IgnoredByGitKind {
			orto.PrintChange(c)
		}
	}
	allDotGitChanges := true
	for _, c := range allChanges {
		if c.Kind == orto.IgnoredByOrtoKind {
			// TODO: fix this:
			if c.FsFile != nil && !strings.HasPrefix(c.FsFile.CleanPath, ".git/") {
				allDotGitChanges = false
				break
			}
		}
	}
	if allDotGitChanges {
		println("⛔︎ OrtoIgnored", ".git/**")
	} else {
		for _, c := range allChanges {
			if c.Kind == orto.IgnoredByOrtoKind {
				orto.PrintChange(c)
			}
		}
	}

	println("---")

	for _, c := range allChanges {
		fmt.Printf("%#v,%#v\n", c.FsFile, c.GitFile)
		switch c.Kind {
		case orto.AddedKind, orto.ModifiedKind:
			if c.FsFile != nil {
				orto.CopyFile(c.FsFile.CleanPath, filepath.Join(params.Destination, c.FsFile.CleanPath))
				orto.PrintCopy(c.FsFile.CleanPath, filepath.Join(params.Destination, c.FsFile.CleanPath))
			}
		case orto.DeletedKind:
			if c.GitFile != nil {
				// TODO: actually do something with deletions
				orto.PrintDel(c.GitFile.CleanPath)
			}
		}
	}

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
