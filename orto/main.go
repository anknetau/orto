package orto

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

func CleanFilePath(path string) string {
	inputParts := SplitFilePath(path)
	temp := make([]string, 0, len(inputParts))
	inRange := func(i int) bool {
		return i < len(inputParts)
	}
	separator := func(i int) bool {
		return inRange(i) && inputParts[i][0] == filepath.Separator
	}
	is := func(i int, s string) bool {
		return inRange(i) && inputParts[i] == s
	}
	for i := 0; i < len(inputParts); i++ {
		s := inputParts[i]
		if s == "." && separator(i+1) {
			i++
			continue
		}
		if !separator(i) && !is(i, "..") && separator(i+1) && is(i+2, "..") {
			if separator(i + 3) {
				i += 3
			} else {
				i += 2
			}
			continue
		}
		if separator(i) {
			temp = append(temp, string(filepath.Separator))
		} else {
			temp = append(temp, s)
		}
	}
	inputParts = temp
	if len(inputParts) == 0 {
		return "."
	}
	// Remove / at the end
	if inputParts[len(inputParts)-1] == "/" && len(inputParts) > 1 {
		inputParts = inputParts[:len(inputParts)-1]
	}
	// Remove /. at the end
	if inputParts[len(inputParts)-1] == "." && len(inputParts) > 2 && inputParts[len(inputParts)-2] == "/" {
		inputParts = inputParts[:len(temp)-2]
	}
	return strings.Join(inputParts, "")
}

// Split a dirty file path, preserving it
// "" --> []
// "aa//aaa" --> ["aa", "//", "aaa"]
// "/aaaa" --> ["/", "aaa"]
// "/////" --> ["////"]
func SplitFilePath(path string) []string {
	lastSeparatorIndex := -1
	var result []string
	for i := 0; i < len(path); i++ {
		if path[i] != filepath.Separator {
			continue
		}
		count := countSeparators(path, i)
		if lastSeparatorIndex != -1 {
			result = append(result, path[lastSeparatorIndex+1:i])
		} else if i != 0 {
			result = append(result, path[0:i])
		}
		result = append(result, strings.Repeat(string(filepath.Separator), count))
		i += count - 1
		lastSeparatorIndex = i
	}
	if lastSeparatorIndex < len(path)-1 {
		// This works with lastSeparatorIndex==-1
		result = append(result, path[lastSeparatorIndex+1:])
	}
	return result
}

func countSeparators(path string, index int) int {
	count := 1
	for ; index+count < len(path) && path[index+count] == filepath.Separator; count++ {
	}
	return count
}

func ValidFilePath(path string) bool {
	separator := string(filepath.Separator)
	sep := "[" + separator + "]"
	// parts := SplitFilePath(path)

	//    /?a+(/|/a+)
	//    /?a or /?a/ or /?a/a
	//    a or /a or a/ or /a/ or a/a or /a/a
	chars := `[a-zA-Z0-9_\-.]`

	re := regexp.MustCompile(sep + "?" + chars + "+)(" + sep + chars + "+)*/?$")
	matches := re.FindStringSubmatch(path)
	if matches == nil {
		return false
	}
	return true
}

func debug(value any) {
	// fmt.Printf("%#v\n", entries)
	b, err := json.MarshalIndent(value, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	println(string(b))
}
