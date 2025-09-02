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

func CheckAndUpdateParameters(params *Parameters) {
	CheckSourceDirectory(params.Source)
	CheckDestinationDirectory(params.Destination)
	params.Source = filepath.Clean(params.Source)
	params.Destination = filepath.Clean(params.Destination)

	// TODO: check this properly:
	if len(params.ChangeSetName) == 0 || strings.ContainsAny(params.ChangeSetName, string(filepath.Separator)+" ") {
		log.Fatalf("Invalid ChangeSetName %s", params.ChangeSetName)
	}
	params.Destination = filepath.Join(params.Destination, params.ChangeSetName) // TODO: do this properly
	// TODO: do this properly:
	err := os.Mkdir(params.Destination, 0755)
	if err != nil {
		log.Fatal(err)
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

// CheckDestinationDirectory git status --porcelain=v2 --untracked-files=all --show-stash --branch --ignored
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
