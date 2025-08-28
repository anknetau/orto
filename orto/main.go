package orto

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Index[T any](fsFiles []T, w func(T) string) map[string]T {
	fsFileIndex := make(map[string]T, len(fsFiles))
	for _, fsFile := range fsFiles {
		fsFileIndex[w(fsFile)] = fsFile
	}
	return fsFileIndex
}

// git status --porcelain=v2 --untracked-files=all --show-stash --branch --ignored
func CheckDestination(path string) {
	destinationStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !destinationStat.IsDir() {
		log.Fatalf("Destination %s is not a directory", path)
	}
}

func CopyFile(src string, dest string) int64 {
	read, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer func(r *os.File) {
		err := r.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(read)
	//os.MkdirAll(dest, 0755)
	// TODO: do this properly, eg make sure we are not going up a level
	dir, _ := filepath.Split(dest)
	//println(dir)
	//println(fileName)
	//os.Exit(1)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	write, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer func(w *os.File) {
		err := w.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(write)
	n, err := write.ReadFrom(read)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func ChangePrint(change Change) {
	switch c := change.(type) {
	case Added:
		println("❇️ Added", c.FsFile.Filepath)
	case Deleted:
		println("❌ Deleted", c.GitFile.Filepath)
	case Unchanged:
		println("➖ Unchanged", c.FsFile.Filepath)
	case Modified:
		println("✏️ Modified", c.FsFile.Filepath)
	case IgnoredByGit:
		println("⛔︎ GitIgnored", c.FsFile.Filepath)
	case IgnoredByOrto:
		println("⛔︎ OrtoIgnored", c.FsFile.Filepath)
	}
}

type Both struct {
	FsFile  FSFile
	GitFile GitFile
}

func CompareFiles(fsFiles []FSFile, gitFiles []GitFile, fsFileIndex map[string]FSFile, gitFileIndex map[string]GitFile) ([]Both, []FSFile, []GitFile) {
	var common []Both
	var left []FSFile
	var right []GitFile
	for _, fsFile := range fsFiles {
		gitFile, ok := gitFileIndex[fsFile.Filepath]
		if ok {
			common = append(common, Both{FsFile: fsFile, GitFile: gitFile})
		} else {
			left = append(left, fsFile)
		}
	}

	for _, gitFile := range gitFiles {
		_, ok := fsFileIndex[gitFile.Filepath]
		if !ok {
			right = append(right, gitFile)
		}
	}
	return common, left, right
}

// TODO: symlinks appear as blobs but with a different mode. also check symlinks on the filesystem.
// TODO: empty dirs?
// TODO: how do we know if a file is the same file? inodes, etc?
// TODO: case change
// TODO: test in windows and linux

func ortoShouldIgnore(fsFile *FSFile) bool {
	if strings.HasPrefix(fsFile.Filepath, ".git"+string(filepath.Separator)) {
		return true
	}
	//if len(parts) > 0 && parts[0] == ".venv" {
	//	return true
	//}
	//if len(parts) > 0 && parts[0] == "out" {
	//	return true
	//}
	//if len(parts) > 0 && parts[0] == "lib" {
	//	return true
	//}
	//if len(parts) > 0 && parts[len(parts)-1] == ".DS_Store" {
	//	return true
	//}
	return false
}

func ComparePair(gitFile *GitFile, fsFile *FSFile, gitIgnoredFilesIndex map[string]string) Change {
	if fsFile != nil {
		if ortoShouldIgnore(fsFile) {
			return IgnoredByOrto{fsFile}
		}
		if _, ignored := gitIgnoredFilesIndex[fsFile.Filepath]; ignored {
			return IgnoredByGit{fsFile}
		}
	}
	if gitFile == nil && fsFile == nil {
		panic("Illegal state")
	}
	if gitFile != nil && fsFile != nil {
		if gitFile.Filepath != fsFile.Filepath {
			panic("Illegal state: " + gitFile.Filepath + " " + fsFile.Filepath)
		}
		stat, err := os.Stat(fsFile.root)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			panic("was dir: " + fsFile.root)
		}
		calculatedChecksum := checksumBlob(gitFile.path, checksumGetAlgo(gitFile.checksum))
		if calculatedChecksum == gitFile.checksum {
			return Unchanged{fsFile, gitFile}
		} else {
			return Modified{fsFile, gitFile}
		}
	} else if gitFile != nil {
		return Deleted{gitFile}
	} else {
		return Added{fsFile}
	}
}

func debug(value any) {
	// fmt.Printf("%#v\n", entries)
	b, err := json.MarshalIndent(value, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	println(string(b))
}
