package orto

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/git"
)

// TODO: symlinks appear as blobs but with a different mode. also check symlinks on the filesystem.
// TODO: empty dirs?
// TODO: how do we know if a file is the same file? inodes, etc?
// TODO: case change
// TODO: test in windows and linux

type Status struct {
	params               Parameters
	fsFiles              []FSFile
	gitBlobs             []git.Blob
	fsFileIndex          map[string]FSFile
	gitFileIndex         map[string]git.Blob
	gitStatus            []git.StatusLine
	gitIgnoredFilesIndex map[string]string
}

func Start(params Parameters) {
	CheckAndUpdateParameters(&params)
	status := gatherFiles(params)
	allChanges := compareFiles(status, params)
	write(status, params, allChanges)
}

func gatherFiles(params Parameters) Status {
	PrintLogHeader("Gathering files...")
	// TODO: check for git version on startup, etc.
	err := os.Chdir(params.Source)
	if err != nil {
		log.Fatal(err)
	}
	status := Status{
		params:    params,
		fsFiles:   FsReadDir(params.Source),
		gitBlobs:  git.RunGetTreeForHead(),
		gitStatus: git.RunStatus(),
	}
	status.fsFileIndex = Index(status.fsFiles, func(file FSFile) string {
		return file.CleanPath
	})
	status.gitFileIndex = Index(status.gitBlobs, func(file git.Blob) string {
		return file.CleanPath
	})

	var gitIgnoredFiles = Filter(status.gitStatus, func(statusLine *git.StatusLine) *string {
		if val, ok := (*statusLine).(git.IgnoredStatusLine); ok {
			return &val.Path
		}
		return nil
	})
	status.gitIgnoredFilesIndex = Index(gitIgnoredFiles, func(s string) string { return s })
	return status
}

func compareFiles(status Status, params Parameters) []Change {
	PrintLogHeader("Comparing...")
	common, left, right := CompareFiles(status.fsFiles, status.gitBlobs, status.fsFileIndex, status.gitFileIndex)

	// Filter out ignored
	//var aLeft []FSFile
	//for _, f := range left {
	//	if _, found := gitIgnoredFilesIndex[f.filepath]; !found {
	//		aLeft = append(aLeft, f)
	//	}
	//}

	var allChanges []Change
	for _, f := range left {
		change := ComparePair(nil, &f, status.gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}
	for _, f := range right {
		change := ComparePair(&f, nil, status.gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}
	for _, c := range common {
		change := ComparePair(&c.GitFile, &c.FsFile, status.gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}

	for _, c := range allChanges {
		if c.Kind == AddedKind || c.Kind == ModifiedKind || c.Kind == DeletedKind {
			PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == UnchangedKind {
			PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == IgnoredByGitKind {
			PrintChange(c)
		}
	}

	ortoIgnores := 0
	ortoDotGitIgnores := 0
	for _, c := range allChanges {
		if c.Kind == IgnoredByOrtoKind {
			ortoIgnores++
			// TODO: fix this:
			if c.FsFile != nil && !strings.HasPrefix(c.FsFile.CleanPath, ".git/") {
				ortoDotGitIgnores++
			}
		}
	}
	if ortoIgnores == ortoDotGitIgnores && ortoIgnores > 0 {
		println("⛔︎ OrtoIgnored", ".git/**")
	} else {
		for _, c := range allChanges {
			if c.Kind == IgnoredByOrtoKind {
				PrintChange(c)
			}
		}
	}

	return allChanges
}

func write(_ Status, params Parameters, allChanges []Change) {
	PrintLogHeader("Writing output...")
	for _, c := range allChanges {
		//fmt.Printf("%#v,%#v\n", c.FsFile, c.Blob)
		switch c.Kind {
		case AddedKind, ModifiedKind:
			if c.FsFile != nil { // TODO: why checking this here?
				CopyFile(c.FsFile.CleanPath, c.FsFile.CleanPath, params.Destination)
				PrintLogCopy(c.FsFile.CleanPath, filepath.Join(params.Destination, c.FsFile.CleanPath))
			}
		case DeletedKind:
			if c.Blob != nil { // TODO: why checking this here?
				// TODO: actually do something with deletions
				PrintLogDel(c.Blob.CleanPath)
			}
		case UnchangedKind:
			if params.Inclusions.UnchangedFiles {
				CopyFile(c.FsFile.CleanPath, c.FsFile.CleanPath, params.Destination)
				PrintLogCopy(c.FsFile.CleanPath, filepath.Join(params.Destination, c.FsFile.CleanPath))
			}
		case IgnoredByGitKind:
			// TODO
		case IgnoredByOrtoKind:
			// TODO

		}
	}
	PrintLogHeader("Finished")
}
