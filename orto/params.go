package orto

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anknetau/orto/fp"
	"github.com/anknetau/orto/git"
	"github.com/anknetau/orto/util"
)

// UserParameters are parameters set by the user
type UserParameters struct {
	Source              string
	Destination         string
	ChangeSetName       string
	PathToGitBinary     string
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
	setDefaultStringIfEmpty(&params.PathToGitBinary, "git")
}

func applyDefaultsAndCheckParameters(params *UserParameters) Settings {
	params.ApplyDefaults()
	startTime := time.Now()

	absSourceDir := CheckSourceDirectory(params.Source)

	gitEnv := git.Find(params.PathToGitBinary, absSourceDir)
	PrintLogHeader("Found git version " + gitEnv.Version + " with algo " + string(gitEnv.Algo))
	PrintLogHeader("Repository worktree is '" + gitEnv.AbsRoot + "' with .git at '" + gitEnv.AbsGitDir + "'")

	absDestinationDir := CheckDestinationDirectory(params.Destination)
	PrintLogHeader("Destination is '" + absDestinationDir + "'")

	// TODO: this is unsupported for now, but will change in the future - if eg the target is a compressed file
	if !fp.AbsolutePathsAreUnrelated(gitEnv.AbsRoot, absDestinationDir) {
		log.Fatalf("Source and destination are related: %s and %s", params.Source, params.Destination)
	}
	if len(params.ChangeSetName) == 0 {
		params.ChangeSetName = util.SerializedDateTime(startTime)
	} else {
		// TODO: check this properly:
		if strings.ContainsAny(params.ChangeSetName, string(filepath.Separator)+" ") {
			log.Fatalf("Invalid ChangeSetName %s", params.ChangeSetName)
		}
	}
	return Settings{
		input: InputSettings{
			copyDotGit: params.CopyDotGit,
		},
		output: OutputSettings{
			absDestinationDir:               absDestinationDir,
			absDestinationChangeSetJsonFile: filepath.Join(absDestinationDir, params.ChangeSetName+".json"),
			absDestinationChangeSetDir:      filepath.Join(absDestinationDir, params.ChangeSetName),
			copyUnchangedFiles:              params.CopyUnchangedFiles,
		},
		envConfig: fp.EnvConfig{
			StartTime: startTime,
		},
		gitEnv: gitEnv,
	}
}

func CheckSourceDirectory(path string) string {
	if len(path) == 0 {
		log.Fatalf("Source '%s' is not a directory", path)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: os.Stat follows symlinks apparently
	sourceStat, err := os.Stat(absPath)
	if err != nil {
		log.Fatal(err)
	}
	if !sourceStat.IsDir() {
		log.Fatalf("Source '%s' is not a directory", path)
	}
	return absPath
}

func CheckDestinationDirectory(path string) string {
	if len(path) == 0 {
		log.Fatalf("Destination '%s' is not a directory", path)
	}
	absDestinationDir, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: os.Stat follows symlinks apparently
	destinationStat, err := os.Stat(absDestinationDir)
	if err != nil {
		log.Fatal(err)
	}
	if !destinationStat.IsDir() {
		log.Fatalf("Destination '%s' is not a directory", path)
	}
	isDirEmpty, err := fp.IsDirEmpty(absDestinationDir)
	if err != nil {
		log.Fatal(err)
	}
	if !isDirEmpty {
		log.Fatalf("Destination %s is not empty", path)
	}
	return absDestinationDir
}

// TODO: destination shouldn't be in source etc
