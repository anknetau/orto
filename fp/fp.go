package fp

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func CleanFilePath(path string) string {
	inputParts := SplitFilePath(path)
	temp := make([]string, 0, len(inputParts))
	inRange := func(i int) bool {
		return i < len(inputParts)
	}
	separator := func(i int) bool {
		return inRange(i) && os.IsPathSeparator(inputParts[i][0])
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
	if len(inputParts) > 1 && os.IsPathSeparator(inputParts[len(inputParts)-1][0]) {
		inputParts = inputParts[:len(inputParts)-1]
	}
	// Remove /. at the end
	//if inputParts[len(inputParts)-1] == "." && len(inputParts) > 2 && inputParts[len(inputParts)-2] == "/" {
	if inputParts[len(inputParts)-1] == "." && len(inputParts) > 2 && os.IsPathSeparator(inputParts[len(inputParts)-2][0]) {
		inputParts = inputParts[:len(inputParts)-2]
	}
	return strings.Join(inputParts, "")
}

// SplitFilePath
// Split a dirty file path, preserving it
// "" --> []
// "aa//aaa" --> ["aa", "//", "aaa"]
// "/aaaa" --> ["/", "aaa"]
// "/////" --> ["/////"]
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
		// This works when lastSeparatorIndex is -1
		result = append(result, path[lastSeparatorIndex+1:])
	}
	if result == nil {
		return []string{}
	}
	return result
}

func ValidFilePathForOrtoOrDie(path string) {
	if !ValidFilePathForOrto(path) {
		log.Fatal("Cannot handle this path: " + path)
	}
}

var (
	reValidFilePathName = regexp.MustCompile(`^[a-zA-Z0-9_.\-~@#$%^&=+{}\[\]:;,<>()]+$`)
)

func ValidFilePathForOrto(path string) bool {
	if len(path) == 0 {
		return false
	}
	parts := SplitFilePath(filepath.Clean(path))
	for _, part := range parts {
		if part[0] != filepath.Separator {
			matches := reValidFilePathName.FindStringSubmatch(part)
			if matches == nil {
				return false
			}
		}
	}
	return true
}

func countSeparators(path string, index int) int {
	count := 1
	for ; index+count < len(path) && path[index+count] == filepath.Separator; count++ {
	}
	return count
}

// DirFirstEntry Get just the first file within the given directory path.
func DirFirstEntry(path string) (*string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entry, err := f.Readdirnames(1)
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(entry) == 0 {
		return nil, nil
	}
	return &entry[0], nil
}

// IsDirEmpty Efficiently determine if the given directory is empty by reading a single entry.
// If passed a non-directory, it will return an error.
func IsDirEmpty(path string) (bool, error) {
	filename, err := DirFirstEntry(path)
	if err != nil {
		return false, err
	}
	if filename == nil {
		return true, nil
	} else {
		return false, err
	}
}

// FilepathParts will return an array of the non-separator parts of the given path.
// For an empty string, it will return {"."}
func FilepathParts(f string) []string {
	return removeSeparators(SplitFilePath(filepath.Clean(f)))
}

func removeSeparators(path []string) []string {
	result := make([]string, 0, len(path))
	for _, v := range path {
		if v[0] != filepath.Separator {
			result = append(result, v)
		}
	}
	return result
}

func AbsolutePathIsParentOrEqual(parent, child string) bool {
	if !filepath.IsAbs(parent) || !filepath.IsAbs(child) {
		log.Fatalf("AbsolutePathIsParentOrEqual: %s and %s are not both absolute", parent, child)
	}
	return startsWith(FilepathParts(child), FilepathParts(parent))
}

func AbsolutePathsAreUnrelated(a, b string) bool {
	return !AbsolutePathIsParentOrEqual(a, b) && !AbsolutePathIsParentOrEqual(b, a)
}

func startsWith[T comparable](s, prefix []T) bool {
	return len(prefix) <= len(s) && slices.Equal(s[:len(prefix)], prefix)
}

func CreateIntermediateDirectoriesForFile(relPathToFileNotDir string, destAbsoluteDirectory string) {
	// TODO: test what happens when only part of the path already exists
	// TODO: test what happens when part of the path exists and is a file and not a directory
	if !filepath.IsAbs(destAbsoluteDirectory) {
		panic("Not an absolute directory: " + destAbsoluteDirectory)
	}
	if filepath.IsAbs(relPathToFileNotDir) {
		panic("Not a relative path: " + relPathToFileNotDir)
	}
	relPathWithoutFilename, _ := filepath.Split(relPathToFileNotDir)
	parts := FilepathParts(relPathWithoutFilename)
	if len(parts) == 0 {
		panic("Invalid state")
	}
	if len(parts) == 1 && parts[0] == "." {
		// No directories to create
		return
	}
	absTargetDir := filepath.Join(destAbsoluteDirectory, relPathWithoutFilename)
	//println("createIntermediateDirectoriesForFile " + relPathToFileNotDir + " to " + destAbsoluteDirectory + " dir=" + absTargetDir + ", relPathWithoutFilename=" + relPathWithoutFilename)
	dirStat, err := os.Stat(absTargetDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(absTargetDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		//println("created " + absTargetDir)
		return
	} else if err != nil {
		log.Fatal(err)
	}
	if !dirStat.IsDir() {
		log.Fatalf("Source %s Already exists and is not a directory", relPathToFileNotDir)
	}
}
