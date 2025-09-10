package orto

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/fp"
	"github.com/anknetau/orto/git"
)

// UserParameters are parameters set by the user
type UserParameters struct {
	Source              string
	Destination         string
	ChangeSetName       string
	GitCommand          string
	CopyDotGit          bool
	CopyGitIgnoredFiles bool // TODO
	CopyUnchangedFiles  bool
	// TODO: CopyContentsOfSubmodules? Do we need to diff those too, recursively?
}

func setDefaultStringIfEmpty(key *string, def string) {
	if *key == "" {
		*key = def
	}
}

func (params *UserParameters) ApplyDefaults() {
	setDefaultStringIfEmpty(&params.GitCommand, "git")
}

func checkAndUpdateUserParameters(params *UserParameters) Settings {
	params.ApplyDefaults()
	gitEnv := git.FindGit(params.GitCommand)
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
	return Settings{
		input: InputSettings{
			absSourceDir: absSourceDir,
			copyDotGit:   params.CopyDotGit,
		},
		output: OutputSettings{
			absDestinationDir:               absDestinationDir,
			absDestinationChangeSetJsonFile: filepath.Join(absDestinationDir, params.ChangeSetName+".json"),
			absDestinationChangeSetDir:      filepath.Join(absDestinationDir, params.ChangeSetName),
			copyUnchangedFiles:              params.CopyUnchangedFiles,
		},
		envConfig: fp.EnvConfig{
			// TODO: move git information into the separate object
			GitCommand: params.GitCommand,
			GitVersion: gitEnv.Version,
		},
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
