package git

import (
	"log"
	"os"
	"path/filepath"

	"github.com/anknetau/orto/fp"
)

type Env struct {
	PathToBinary string
	Version      string
	Algo         fp.Algo
	AbsRoot      string
	AbsGitDir    string
}

func Find(pathToBinary string, absPathToChdir string) Env {
	if !filepath.IsAbs(absPathToChdir) {
		panic("Not an absolute path: " + absPathToChdir)
	}
	err := os.Chdir(absPathToChdir)
	if err != nil {
		log.Fatal(err)
	}
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

	return env
}

func (env Env) IsPartOfDotGit(path string) bool {
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Fatal(err)
		}
		path = absPath
	}
	return fp.AbsolutePathIsParentOrEqual(env.AbsGitDir, path)
}
