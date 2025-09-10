package git

import (
	"log"

	"github.com/anknetau/orto/fp"
)

type Env struct {
	PathToBinary string
	Version      string
	Algo         fp.Algo
	AbsRoot      string
	AbsGitDir    string
}

func Find(pathToBinary string) Env {
	//_ = os.Chdir("/")
	version := RunVersion(pathToBinary)
	if version == "" {
		log.Fatal("Could not find git")
	}
	env := Env{
		PathToBinary: pathToBinary,
		Version:      version,
		Algo:         fp.UNKNOWN,
	}

	worktreeStatus := env.RunGetIsInsideWorktree()
	switch worktreeStatus {
	case WorktreeStatusTrue:
	case WorktreeStatusFalse:
		log.Fatal("Not a inside a working tree")
	case WorktreeStatusNotARepo:
		log.Fatal("Not a git repository")
	}

	absRoot := env.RunGetRepoRoot()
	fp.IsAbsPathToDirOrDie(absRoot, "Repository root")
	env.AbsRoot = absRoot

	absGitDir := env.RunGetGitDir()
	fp.IsAbsPathToDirOrDie(absRoot, "Directory .git")
	env.AbsGitDir = absGitDir

	env.Algo = RunGetRepoHashFormat(pathToBinary)

	// TODO: look for .git in parent directories
	return env
}
