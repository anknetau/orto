package orto

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/fp"
	"github.com/anknetau/orto/git"
)

// TODO: symlinks appear as blobs but with a different mode. also check symlinks on the filesystem.
// TODO: empty dirs?
// TODO: how do we know if a file is the same file? inodes, etc?
// TODO: case change
// TODO: test in windows and linux

// TODO: We have Inputs and Input! Rename one.
type Inputs struct {
	fsFiles              []FSFile
	gitBlobs             []git.Blob
	fsFileIndex          map[string]FSFile
	gitFileIndex         map[string]git.Blob
	gitStatus            []git.StatusLine
	gitIgnoredFilesIndex map[string]string
	envConfig            fp.EnvConfig
}

type Input struct {
	absSourceDir string
	envConfig    fp.EnvConfig
}

type Output struct {
	absDestinationDir               string
	absDestinationChangeSetDir      string
	absDestinationChangeSetJsonFile string
}

func Start(params Parameters) {
	input, output := checkAndUpdateParameters(&params)
	inputs := gatherFiles(input)
	allChanges := compareFiles(inputs)
	write(params.GitCommand, inputs, output, params.Inclusions.UnchangedFiles, allChanges)
}

func gatherFiles(input Input) Inputs {
	absSourceDir := input.absSourceDir
	PrintLogHeader("Found git version " + input.envConfig.GitVersion)
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
		gitBlobs:  git.RunGetTreeForHead(input.envConfig),
		gitStatus: git.RunStatus(input.envConfig),
	}
	inputs.fsFileIndex = Index(inputs.fsFiles, func(file FSFile) string {
		return file.CleanPath
	})
	inputs.gitFileIndex = Index(inputs.gitBlobs, func(file git.Blob) string {
		return file.CleanPath
	})

	var gitIgnoredFiles = Filter(inputs.gitStatus, func(statusLine *git.StatusLine) *string {
		git.PrintStatusLine(statusLine)
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
		if c.Kind == ChangeKindAdded || c.Kind == ChangeKindModified || c.Kind == ChangeKindDeleted {
			PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == ChangeKindUnchanged {
			PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == ChangeKindIgnoredByGit {
			PrintChange(c)
		}
	}

	ortoIgnores := 0
	ortoDotGitIgnores := 0
	for _, c := range allChanges {
		if c.Kind == ChangeKindIgnoredByOrto {
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
			if c.Kind == ChangeKindIgnoredByOrto {
				PrintChange(c)
			}
		}
	}

	return allChanges
}

func write(gitCommand string, _ Inputs, output Output, copyUnchangedFiles bool, allChanges []Change) {
	return
	PrintLogHeader("Writing output...")

	// TODO: do this properly:
	err := os.Mkdir(output.absDestinationChangeSetDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	jsonOut := JsonOutput{absDestinationChangeSetJsonFile: output.absDestinationChangeSetJsonFile}
	defer jsonOut.close()
	jsonOut.start()

	for _, change := range allChanges {
		//fmt.Printf("%#v,%#v\n", change.FsFile, change.Blob)
		validateChange(change)
		switch change.Kind {
		case ChangeKindAdded:
			CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, output.absDestinationChangeSetDir)
			PrintLogCopy(change.FsFile.CleanPath, filepath.Join(output.absDestinationChangeSetDir, change.FsFile.CleanPath))
		case ChangeKindModified:
			// TODO: copy the old file too
			CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, output.absDestinationChangeSetDir)
			PrintLogCopy(change.FsFile.CleanPath, filepath.Join(output.absDestinationChangeSetDir, change.FsFile.CleanPath))
		case ChangeKindDeleted:
			SaveGitBlob(gitCommand, change.Blob.Checksum, change.Blob.CleanPath, output.absDestinationChangeSetDir)
			jsonOut.maybeAddComma()
			jsonOut.encode(change.Blob)
			PrintLogDel(change.Blob.CleanPath)
		case ChangeKindUnchanged:
			if copyUnchangedFiles {
				CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, output.absDestinationChangeSetDir)
				PrintLogCopy(change.FsFile.CleanPath, filepath.Join(output.absDestinationChangeSetDir, change.FsFile.CleanPath))
			}
		case ChangeKindIgnoredByGit:
			// TODO
		case ChangeKindIgnoredByOrto:
			// TODO
		}
	}

	PrintLogHeader("Written " + output.absDestinationChangeSetJsonFile)

	PrintLogHeader("Finished")
}

func validateChange(c Change) {
	switch c.Kind {
	case ChangeKindAdded:
		// Only FsFile
		if c.FsFile == nil || c.Blob != nil {
			panic("Illegal state")
		}
	case ChangeKindDeleted:
		// Only blob
		if c.FsFile != nil || c.Blob == nil {
			panic("Illegal state")
		}
	case ChangeKindUnchanged:
		// Has both
		if c.Blob == nil || c.FsFile == nil {
			panic("Illegal state")
		}
	case ChangeKindModified:
		// Has both
		if c.FsFile == nil || c.Blob == nil {
			panic("Illegal state")
		}
	case ChangeKindIgnoredByGit:
		// Only FSFile
		if c.FsFile == nil || c.Blob != nil {
			panic("Illegal state")
		}
	case ChangeKindIgnoredByOrto:
		// Only FSFile
		if c.FsFile == nil || c.Blob != nil {
			panic("Illegal state")
		}
	}
}
