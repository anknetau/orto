package orto

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/anknetau/orto/fp"
)

func Index[T any](fsFiles []T, w func(T) string) map[string]T {
	fsFileIndex := make(map[string]T, len(fsFiles))
	for _, fsFile := range fsFiles {
		fsFileIndex[w(fsFile)] = fsFile
	}
	return fsFileIndex
}

func CheckSource(path string) {
	sourceStat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !sourceStat.IsDir() {
		log.Fatalf("Source %s is not a directory", path)
	}
	dotGitStat, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		log.Fatal(err)
	}
	if !dotGitStat.IsDir() {
		log.Fatalf("Source %s does not contain a .git subdirectory", path)
	}
}

// CheckDestination git status --porcelain=v2 --untracked-files=all --show-stash --branch --ignored
func CheckDestination(path string) {
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
}

// TODO: destination shouldn't be in source etc

func CopyFile(src string, dest string) int64 {
	// TODO: commented out.
	println(src + " copied to " + dest)
	return 0
	// TODO: this is quick and dirty
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

func PrintCopy(src string, dst string) {
	println(src + " → " + dst)
}

func PrintDel(src string) {
	println(src + " ❌ ")
}

func PrintChange(change Change) {
	switch change.Kind {
	case AddedKind:
		println("❇️ Added", change.FsFile.CleanPath)
	case DeletedKind:
		println("❌ Deleted", change.GitFile.CleanPath)
	case UnchangedKind:
		println("➖ Unchanged", change.FsFile.CleanPath)
	case ModifiedKind:
		println("✏️ Modified", change.FsFile.CleanPath)
	case IgnoredByGitKind:
		println("⛔︎ GitIgnored", change.FsFile.CleanPath)
	case IgnoredByOrtoKind:
		println("⛔︎ OrtoIgnored", change.FsFile.CleanPath)
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
		gitFile, ok := gitFileIndex[fsFile.CleanPath]
		if ok {
			common = append(common, Both{FsFile: fsFile, GitFile: gitFile})
		} else {
			left = append(left, fsFile)
		}
	}

	for _, gitFile := range gitFiles {
		_, ok := fsFileIndex[gitFile.CleanPath]
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

func isOrtoIgnored(fsFile *FSFile, destination string) bool {
	splitParts := fp.SplitFilePath(fp.CleanFilePath(fsFile.CleanPath))
	if len(splitParts) > 0 && splitParts[0] == ".git" {
		return true
	}
	// TODO: this is within the output!
	// TODO: ensure this is the right way of comparing - think absolute vs. rel, etc.
	if len(splitParts) > 0 && splitParts[0] == "dest" || fp.CleanFilePath(fsFile.CleanPath) == destination {
		log.Fatalf("a! Orto ignored file: " + fsFile.CleanPath)
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

func ComparePair(gitFile *GitFile, fsFile *FSFile, gitIgnoredFilesIndex map[string]string, destination string) Change {
	if fsFile != nil {
		if isOrtoIgnored(fsFile, destination) {
			return Change{Kind: IgnoredByOrtoKind, FsFile: fsFile}
		}
		if _, ignored := gitIgnoredFilesIndex[fsFile.CleanPath]; ignored {
			return Change{Kind: IgnoredByGitKind, FsFile: fsFile}
		}
	}
	if gitFile == nil && fsFile == nil {
		panic("Illegal state")
	}
	if gitFile != nil && fsFile != nil {
		if gitFile.CleanPath != fsFile.CleanPath {
			panic("Illegal state: " + gitFile.CleanPath + " " + fsFile.CleanPath)
		}
		stat, err := os.Stat(fsFile.path)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			panic("was dir: " + fsFile.path)
		}
		calculatedChecksum := checksumBlob(gitFile.path, checksumGetAlgo(gitFile.checksum))
		if calculatedChecksum == gitFile.checksum {
			return Change{Kind: UnchangedKind, FsFile: fsFile, GitFile: gitFile}
		} else {
			return Change{Kind: ModifiedKind, FsFile: fsFile, GitFile: gitFile}
		}
	} else if gitFile != nil {
		return Change{Kind: DeletedKind, GitFile: gitFile}
	} else {
		return Change{Kind: AddedKind, FsFile: fsFile}
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

type OrtoParameters struct {
	Source      string
	Destination string
}

func Start(params OrtoParameters) {

}
