package orto

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/fp"
)

type Inclusions struct {
	DotGit          bool // TODO
	GitIgnoredFiles bool // TODO
	UnchangedFiles  bool // TODO
}

type Parameters struct {
	Source        string
	Destination   string
	ChangeSetName string
	Inclusions    Inclusions
}

func checkAndUpdateParameters(params *Parameters) (string, Output) {
	// TODO: check for git version on startup, etc.
	// TODO: Automatically resolve source/dest to be absolute paths
	if !filepath.IsAbs(params.Source) {
		log.Fatal("Source is not an absolute path: " + params.Source)
	}
	if !filepath.IsAbs(params.Destination) {
		log.Fatal("Destination is not an absolute path: " + params.Destination)
	}
	CheckSourceDirectory(params.Source)
	CheckDestinationDirectory(params.Destination)
	absSourceDir, err := filepath.Abs(params.Source)
	if err != nil {
		log.Fatal(err)
	}
	absDestinationDir, err := filepath.Abs(params.Destination)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: this is unsupported for now, but will change in the future - if eg the target is a compressed file
	if !fp.AbsolutePathsAreUnrelated(absSourceDir, absDestinationDir) {
		log.Fatalf("Source and destination are related: %s and %s", params.Source, params.Destination)
	}
	// TODO: check this properly:
	if len(params.ChangeSetName) == 0 || strings.ContainsAny(params.ChangeSetName, string(filepath.Separator)+" ") {
		log.Fatalf("Invalid ChangeSetName %s", params.ChangeSetName)
	}
	return absSourceDir, Output{
		absDestinationDir:               absDestinationDir,
		absDestinationChangeSetJsonFile: filepath.Join(absDestinationDir, params.ChangeSetName+".json"),
		absDestinationChangeSetDir:      filepath.Join(absDestinationDir, params.ChangeSetName),
	}
}

func CheckSourceDirectory(path string) {
	// TODO: os.Stat follows symlinks apparently
	sourceStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !sourceStat.IsDir() {
		log.Fatalf("Source %s is not a directory", path)
	}
	// TODO: os.Stat follows symlinks apparently
	dotGitStat, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		log.Fatal(err)
	}
	if !dotGitStat.IsDir() {
		log.Fatalf("Source %s does not contain a .git subdirectory", path)
	}
	if !filepath.IsAbs(path) {
		log.Fatal("Source is not an absolute path: " + path)
	}
}

func CheckDestinationDirectory(path string) {
	// TODO: os.Stat follows symlinks apparently
	destinationStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !destinationStat.IsDir() {
		log.Fatalf("Destination %s is not a directory", path)
	}
	isDirEmpty, err := fp.IsDirEmpty(path)
	if err != nil {
		log.Fatal(err)
	}
	if !isDirEmpty {
		log.Fatalf("Destination %s is not empty", path)
	}
	if !filepath.IsAbs(path) {
		log.Fatal("Destination is not an absolute path: " + path)
	}
}

// TODO: destination shouldn't be in source etc
