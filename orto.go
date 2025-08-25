package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// git status --porcelain=v2 --untracked-files=all --show-stash --branch --ignored
func main() {
	err := os.Chdir("/Users/ank/dev/accounting/accounting")
	//err := os.Chdir("/Users/ank/dev/mirrors")
	if err != nil {
		log.Fatal(err)
	}
	// FS Files
	fsFiles := fsReadDir("./")
	gitFiles := gitRunGetTreeForHead()
	gitStatus := gitRunStatus()

	println("------")
	for _, status := range gitStatus {
		print(status.Type())
	}
	println("------")

	fsFileIndex := fsIndexByName(fsFiles)
	gitFileIndex := gitIndexByName(gitFiles)

	common, left, right := compareFiles(fsFiles, gitFiles, fsFileIndex, gitFileIndex)
	println("--- common")
	for _, c := range common {
		compare(&c.gitFile, &c.fsFile)
	}
	println("--- left")
	for _, f := range left {
		compare(nil, &f)
	}
	println("--- right")
	for _, f := range right {
		compare(&f, nil)
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

type Both struct {
	fsFile  FSFile
	gitFile GitFile
}

func compareFiles(fsFiles []FSFile, gitFiles []GitFile, fsFileIndex map[string]FSFile, gitFileIndex map[string]GitFile) ([]Both, []FSFile, []GitFile) {
	var common []Both
	var left []FSFile
	var right []GitFile
	for _, fsFile := range fsFiles {
		gitFile, ok := gitFileIndex[fsFile.filepath]
		if ok {
			common = append(common, Both{fsFile: fsFile, gitFile: gitFile})
		} else {
			left = append(left, fsFile)
		}
	}

	for _, gitFile := range gitFiles {
		_, ok := fsFileIndex[gitFile.filepath]
		if !ok {
			right = append(right, gitFile)
		}
	}
	return common, left, right
}

// TODO: symlinks?
// TODO: empty dirs?
// TODO: case change
// TODO: test in windows

func splitFilePath(cleanFilePath string) []string {
	separator := filepath.Join("a", "a")[1]
	return strings.Split(cleanFilePath, string(separator))
}

func shouldIgnore(fsFile *FSFile) bool {
	parts := splitFilePath(fsFile.filepath)
	if len(parts) > 0 && parts[0] == ".git" {
		return true
	}
	if len(parts) > 0 && parts[0] == ".venv" {
		return true
	}
	if len(parts) > 0 && parts[0] == "out" {
		return true
	}
	if len(parts) > 0 && parts[0] == "lib" {
		return true
	}
	if len(parts) > 0 && parts[len(parts)-1] == ".DS_Store" {
		return true
	}
	return false
}

func compare(gitFile *GitFile, fsFile *FSFile) {
	if fsFile != nil && shouldIgnore(fsFile) {
		return
	}
	if gitFile == nil && fsFile == nil {
		panic("Illegal state")
	}
	if gitFile != nil && fsFile != nil {
		if gitFile.filepath != fsFile.filepath {
			panic("Illegal state: " + gitFile.filepath + " " + fsFile.filepath)
		}
		stat, err := os.Stat(fsFile.root)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			panic("was dir: " + fsFile.root)
		}
		calculatedChecksum := checksumBlob(gitFile.path, checksumGetAlgo(gitFile.checksum))
		if calculatedChecksum == gitFile.checksum {
			println("Unchanged: ", gitFile.path)
		} else {
			println("Modified: ", gitFile.path)
		}
	} else if gitFile != nil {
		println("❌  Deleted: ", gitFile.path)
	} else {
		// ❇️ New
		println("New or ignored:", fsFile.root)
	}
}

type StatusLineType int

const (
	IgnoredStatusLineType StatusLineType = iota
	CommentStatusLineType
)

type StatusLine interface {
	Type() StatusLineType
}

type IgnoredStatusLine struct {
	path string
}

type CommentStatusLine struct {
	path string
}

func (IgnoredStatusLine) Type() StatusLineType { return IgnoredStatusLineType }
func (CommentStatusLine) Type() StatusLineType { return CommentStatusLineType }

func debug(value any) {
	// fmt.Printf("%#v\n", entries)
	b, err := json.MarshalIndent(value, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	println(string(b))
}
