package orto

import (
	"log"
	"os"
	"path/filepath"

	"github.com/anknetau/orto/fp"
	"github.com/anknetau/orto/git"
)

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
	copyDotGit   bool
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
	changes := diff(catalog, settings.input)
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

func diff(catalog Catalog, inputSettings InputSettings) []Change {
	PrintLogHeader("Comparing...")
	common, fsFiles, gitBlobs := CompareFiles(catalog.gitBlobs, catalog.fsFiles, catalog.fsFileIndex, catalog.gitBlobIndex)

	// TODO: is this happening or not?
	// Filter out ignored
	//var aLeft []FSFile
	//for _, fsFile := range fsFiles {
	//	if _, found := gitIgnoredFilesIndex[fsFile.filepath]; !found {
	//		aLeft = append(aLeft, fsFile)
	//	}
	//}

	var changes []Change
	for _, fsFile := range fsFiles {
		change := ComparePair(nil, &fsFile, catalog.gitIgnoredFilesIndex, inputSettings)
		changes = append(changes, change)
	}
	for _, blob := range gitBlobs {
		change := ComparePair(&blob, nil, catalog.gitIgnoredFilesIndex, inputSettings)
		changes = append(changes, change)
	}
	for _, gitBlobAndFile := range common {
		change := ComparePair(&gitBlobAndFile.GitBlob, &gitBlobAndFile.FsFile, catalog.gitIgnoredFilesIndex, inputSettings)
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

	ortoDotGitIgnores := 0
	for _, c := range changes {
		validateChange(c)
		// TODO: should ignored files by orto refer just to FsFiles? can i not ignore files in git?
		// or am i ignoring changes?
		if c.Kind == ChangeKindIgnoredByOrto {
			// TODO: this is repeated:
			splitParts := fp.SplitFilePath(c.FsFile.CleanPath)
			if len(splitParts) > 0 && splitParts[0] == ".git" {
				ortoDotGitIgnores++
			}
		}
	}
	if ortoDotGitIgnores > 0 && !inputSettings.copyDotGit {
		println("⛔︎ OrtoIgnored", ".git/**")
	}
	for _, c := range changes {
		if c.Kind == ChangeKindIgnoredByOrto {
			PrintChange(c)
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
