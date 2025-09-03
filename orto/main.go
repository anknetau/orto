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

type Inputs struct {
	fsFiles              []FSFile
	gitBlobs             []git.Blob
	fsFileIndex          map[string]FSFile
	gitFileIndex         map[string]git.Blob
	gitStatus            []git.StatusLine
	gitIgnoredFilesIndex map[string]string
}

type Output struct {
	absDestinationDir          string
	absDestinationChangeSetDir string
}

func Start(params Parameters) {
	absSourceDir, output := checkAndUpdateParameters(&params)
	inputs := gatherFiles(absSourceDir)
	allChanges := compareFiles(inputs)
	write(inputs, output, params.Inclusions.UnchangedFiles, allChanges)
}

func gatherFiles(absSourceDir string) Inputs {
	if !filepath.IsAbs(absSourceDir) {
		panic("Not an absolute directory: " + absSourceDir)
	}
	PrintLogHeader("Gathering files...")
	err := os.Chdir(absSourceDir)
	if err != nil {
		log.Fatal(err)
	}
	inputs := Inputs{
		fsFiles:   FsReadDir(absSourceDir),
		gitBlobs:  git.RunGetTreeForHead(),
		gitStatus: git.RunStatus(),
	}
	inputs.fsFileIndex = Index(inputs.fsFiles, func(file FSFile) string {
		return file.CleanPath
	})
	inputs.gitFileIndex = Index(inputs.gitBlobs, func(file git.Blob) string {
		return file.CleanPath
	})

	var gitIgnoredFiles = Filter(inputs.gitStatus, func(statusLine *git.StatusLine) *string {
		if val, ok := (*statusLine).(git.IgnoredStatusLine); ok {
			return &val.Path
		}
		return nil
	})
	inputs.gitIgnoredFilesIndex = Index(gitIgnoredFiles, func(s string) string { return s })
	return inputs
}

func compareFiles(status Inputs) []Change {
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
		change := ComparePair(nil, &f, status.gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}
	for _, f := range right {
		change := ComparePair(&f, nil, status.gitIgnoredFilesIndex)
		allChanges = append(allChanges, change)
	}
	for _, c := range common {
		change := ComparePair(&c.GitFile, &c.FsFile, status.gitIgnoredFilesIndex)
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

func write(_ Inputs, output Output, copyUnchangedFiles bool, allChanges []Change) {
	PrintLogHeader("Writing output...")

	// TODO: do this properly:
	err := os.Mkdir(output.absDestinationChangeSetDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	for _, change := range allChanges {
		//fmt.Printf("%#v,%#v\n", change.FsFile, change.Blob)
		validateChange(change)
		switch change.Kind {
		case AddedKind:
			CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, output.absDestinationChangeSetDir)
			PrintLogCopy(change.FsFile.CleanPath, filepath.Join(output.absDestinationChangeSetDir, change.FsFile.CleanPath))
		case ModifiedKind:
			// TODO: copy the old file too
			CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, output.absDestinationChangeSetDir)
			PrintLogCopy(change.FsFile.CleanPath, filepath.Join(output.absDestinationChangeSetDir, change.FsFile.CleanPath))
		case DeletedKind:
			// TODO: copy the deleted file?
			// TODO: save the "deletion" somewhere
			PrintLogDel(change.Blob.CleanPath)
		case UnchangedKind:
			if copyUnchangedFiles {
				CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, output.absDestinationChangeSetDir)
				PrintLogCopy(change.FsFile.CleanPath, filepath.Join(output.absDestinationChangeSetDir, change.FsFile.CleanPath))
			}
		case IgnoredByGitKind:
			// TODO
		case IgnoredByOrtoKind:
			// TODO

		}
	}
	PrintLogHeader("Finished")
}

func validateChange(c Change) {
	switch c.Kind {
	case AddedKind:
		// Only FsFile
		if c.FsFile == nil || c.Blob != nil {
			panic("Illegal state")
		}
	case DeletedKind:
		// Only blob
		if c.FsFile != nil || c.Blob == nil {
			panic("Illegal state")
		}
	case UnchangedKind:
		// Has both
		if c.Blob == nil || c.FsFile == nil {
			panic("Illegal state")
		}
	case ModifiedKind:
		// Has both
		if c.FsFile == nil || c.Blob == nil {
			panic("Illegal state")
		}
	case IgnoredByGitKind:
		// Only fsfile
		if c.FsFile == nil || c.Blob != nil {
			panic("Illegal state")
		}
	case IgnoredByOrtoKind:
		// Only fsfile
		if c.FsFile == nil || c.Blob != nil {
			panic("Illegal state")
		}
	}
}
