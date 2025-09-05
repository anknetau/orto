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
// TODO: figure out if there's been a case change, think about what that can do.
// TODO: test in windows and linux

type Catalog struct {
	fsFiles              []FSFile
	gitBlobs             []git.Blob
	gitStatus            []git.StatusLine
	gitSubmodules        []git.Submodule
	fsFileIndex          map[string]FSFile
	gitBlobIndex         map[string]git.Blob
	gitIgnoredFilesIndex map[string]string
	envConfig            fp.EnvConfig
}

type Settings struct {
	input     InputSettings
	output    OutputSettings
	envConfig fp.EnvConfig
}
type InputSettings struct {
	absSourceDir string
}

type OutputSettings struct {
	absDestinationDir               string
	absDestinationChangeSetDir      string
	absDestinationChangeSetJsonFile string
	copyUnchangedFiles              bool
}

func Run(params UserParameters) {
	settings := checkAndUpdateUserParameters(&params)
	catalog := find(settings.input, settings.envConfig)
	changes := diff(catalog)
	write(settings.envConfig, settings.output, changes)
}

func find(inputSettings InputSettings, envConfig fp.EnvConfig) Catalog {
	absSourceDir := inputSettings.absSourceDir
	PrintLogHeader("Found git version " + envConfig.GitVersion)
	if !filepath.IsAbs(absSourceDir) {
		panic("Not an absolute directory: " + absSourceDir)
	}
	PrintLogHeader("Gathering files...")
	err := os.Chdir(absSourceDir)
	if err != nil {
		log.Fatal(err)
	}
	gitBlobs, gitSubmodules := git.RunGetTreeForHead(envConfig)
	inputs := Catalog{
		fsFiles:       FsReadDir(absSourceDir),
		gitBlobs:      gitBlobs,
		gitStatus:     git.RunStatus(envConfig),
		gitSubmodules: gitSubmodules,
	}
	inputs.fsFileIndex = Index(inputs.fsFiles, func(file FSFile) string {
		return file.CleanPath
	})
	inputs.gitBlobIndex = Index(inputs.gitBlobs, func(file git.Blob) string {
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

func diff(catalog Catalog) []Change {
	PrintLogHeader("Comparing...")
	common, left, right := CompareFiles(catalog.fsFiles, catalog.gitBlobs, catalog.fsFileIndex, catalog.gitBlobIndex)

	// TODO: is this happening or not?
	// Filter out ignored
	//var aLeft []FSFile
	//for _, f := range left {
	//	if _, found := gitIgnoredFilesIndex[f.filepath]; !found {
	//		aLeft = append(aLeft, f)
	//	}
	//}

	var changes []Change
	for _, f := range left {
		change := ComparePair(nil, &f, catalog.gitIgnoredFilesIndex)
		changes = append(changes, change)
	}
	for _, f := range right {
		change := ComparePair(&f, nil, catalog.gitIgnoredFilesIndex)
		changes = append(changes, change)
	}
	for _, c := range common {
		change := ComparePair(&c.GitBlob, &c.FsFile, catalog.gitIgnoredFilesIndex)
		changes = append(changes, change)
	}

	for _, c := range changes {
		if c.Kind == ChangeKindAdded || c.Kind == ChangeKindModified || c.Kind == ChangeKindDeleted {
			PrintChange(c)
		}
	}
	for _, c := range changes {
		if c.Kind == ChangeKindUnchanged {
			PrintChange(c)
		}
	}
	for _, c := range changes {
		if c.Kind == ChangeKindIgnoredByGit {
			PrintChange(c)
		}
	}

	ortoIgnores := 0
	ortoDotGitIgnores := 0
	for _, c := range changes {
		validateChange(c)
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
		for _, c := range changes {
			if c.Kind == ChangeKindIgnoredByOrto {
				PrintChange(c)
			}
		}
	}

	return changes
}

func write(envConfig fp.EnvConfig, outputSettings OutputSettings, changes []Change) {
	PrintLogHeader("Writing output...")

	// TODO: do this properly:
	err := os.Mkdir(outputSettings.absDestinationChangeSetDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	jsonOut := JsonOutput{absDestinationChangeSetJsonFile: outputSettings.absDestinationChangeSetJsonFile}
	defer jsonOut.close()
	jsonOut.start()

	for _, change := range changes {
		//fmt.Printf("%#v,%#v\n", change.FsFile, change.GitBlob)
		switch change.Kind {
		case ChangeKindAdded:
			CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, outputSettings.absDestinationChangeSetDir)
			PrintLogCopy(change.FsFile.CleanPath, filepath.Join(outputSettings.absDestinationChangeSetDir, change.FsFile.CleanPath))
		case ChangeKindModified:
			// TODO: copy the old file too
			CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, outputSettings.absDestinationChangeSetDir)
			PrintLogCopy(change.FsFile.CleanPath, filepath.Join(outputSettings.absDestinationChangeSetDir, change.FsFile.CleanPath))
		case ChangeKindDeleted:
			SaveGitBlob(envConfig.GitCommand, change.GitBlob.Checksum, change.GitBlob.CleanPath, outputSettings.absDestinationChangeSetDir)
			jsonOut.maybeAddComma() // TODO: finish this
			jsonOut.encode(change.GitBlob)
			PrintLogDel(change.GitBlob.CleanPath)
		case ChangeKindUnchanged:
			if outputSettings.copyUnchangedFiles {
				CopyFile(change.FsFile.CleanPath, change.FsFile.CleanPath, outputSettings.absDestinationChangeSetDir)
				PrintLogCopy(change.FsFile.CleanPath, filepath.Join(outputSettings.absDestinationChangeSetDir, change.FsFile.CleanPath))
			}
		case ChangeKindIgnoredByGit:
			// TODO
		case ChangeKindIgnoredByOrto:
			// TODO
		}
	}

	PrintLogHeader("Written " + outputSettings.absDestinationChangeSetJsonFile)

	PrintLogHeader("Finished")
}

func validateChange(c Change) {
	switch c.Kind {
	case ChangeKindAdded:
		// Only FsFile
		if c.FsFile == nil || c.GitBlob != nil {
			panic("Illegal state")
		}
	case ChangeKindDeleted:
		// Only blob
		if c.FsFile != nil || c.GitBlob == nil {
			panic("Illegal state")
		}
	case ChangeKindUnchanged:
		// Has both
		if c.GitBlob == nil || c.FsFile == nil {
			panic("Illegal state")
		}
	case ChangeKindModified:
		// Has both
		if c.FsFile == nil || c.GitBlob == nil {
			panic("Illegal state")
		}
	case ChangeKindIgnoredByGit:
		// Only FSFile
		if c.FsFile == nil || c.GitBlob != nil {
			panic("Illegal state")
		}
	case ChangeKindIgnoredByOrto:
		// Only FSFile
		if c.FsFile == nil || c.GitBlob != nil {
			panic("Illegal state")
		}
	}
}
