package git

import (
	"log"

	"github.com/anknetau/orto/fp"
)

type GitEnv struct {
	Command string
	Version string
	Algo    fp.Algo
}

func FindGit(gitCommand string) GitEnv {
	gitVersion := RunVersion(gitCommand)
	if gitVersion == "" {
		log.Fatal("Could not find git")
	}
	gitEnv := GitEnv{
		Command: gitCommand,
		Version: gitVersion,
		Algo:    RunGetRepoHashFormat(gitCommand),
	}

	log.Printf("Git repo checksum format: %s\n", string(gitEnv.Algo))
	// TODO: look for .git in parent directories
	return gitEnv
}
