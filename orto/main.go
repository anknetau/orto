package orto

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anknetau/orto/fp"
	"github.com/anknetau/orto/git"
)

func Index[T any](items []T, callback func(T) string) map[string]T {
	fsFileIndex := make(map[string]T, len(items))
	for _, fsFile := range items {
		fsFileIndex[callback(fsFile)] = fsFile
	}
	return fsFileIndex
}

func CheckSource(path string) {
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

// CheckDestination git status --porcelain=v2 --untracked-files=all --show-stash --branch --ignored
func CheckDestination(path string) {
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

func copyContents(src *os.File, dest *os.File) int64 {
	n, err := dest.ReadFrom(src)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func CopyFile(sourceRelativePath string, destRelativePath string, destAbsoluteDirectory string) int64 {
	//println(sourceRelativePath + " copied to " + destRelativePath + " in " + destAbsoluteDirectory)
	if !filepath.IsAbs(destAbsoluteDirectory) {
		panic("Not an absolute directory: " + destAbsoluteDirectory)
	}
	if !filepath.IsLocal(sourceRelativePath) {
		panic("Non-local source directory " + sourceRelativePath)
	}
	if !filepath.IsLocal(destRelativePath) {
		panic("Non-local destRelativePath directory " + sourceRelativePath)
	}
	destAbsoluteFile := filepath.Join(destAbsoluteDirectory, destRelativePath)

	// TODO: this is quick and dirty
	read, err := os.Open(sourceRelativePath)
	if err != nil {
		log.Fatal(err)
	}
	defer read.Close()

	// TODO: do this properly, eg make sure we are not going up a level
	// TODO: we are assuming here that this is a file and not a directory.
	// If that happens, then we are further assuming what's left is a directory.
	dir, _ := filepath.Split(destRelativePath)
	if len(dir) > 0 {
		dirToCreate := filepath.Join(destAbsoluteDirectory, dir)
		//println("Creating '" + dir + "' at [" + dirToCreate + "] '" + fn + "' for " + destRelativePath)
		err = os.MkdirAll(dirToCreate, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	write, err := os.Create(destAbsoluteFile)
	if err != nil {
		log.Fatal(err)
	}
	defer write.Close()
	return copyContents(read, write)
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
		println("❌ Deleted", change.Blob.CleanPath)
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
	GitFile git.Blob
}

func CompareFiles(fsFiles []FSFile, gitBlobs []git.Blob, fsFileIndex map[string]FSFile, gitFileIndex map[string]git.Blob) ([]Both, []FSFile, []git.Blob) {
	var common []Both
	var left []FSFile
	var right []git.Blob
	for _, fsFile := range fsFiles {
		gitFile, ok := gitFileIndex[fsFile.CleanPath]
		if ok {
			common = append(common, Both{FsFile: fsFile, GitFile: gitFile})
		} else {
			left = append(left, fsFile)
		}
	}

	for _, gitFile := range gitBlobs {
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

func ComparePair(gitFile *git.Blob, fsFile *FSFile, gitIgnoredFilesIndex map[string]string, destination string) Change {
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
		// TODO: os.Stat follows symlinks apparently
		stat, err := os.Stat(fsFile.Path)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			panic("was dir: " + fsFile.Path)
		}
		calculatedChecksum := fp.ChecksumBlob(gitFile.Path, fp.ChecksumGetAlgo(gitFile.Checksum))
		if calculatedChecksum == gitFile.Checksum {
			return Change{Kind: UnchangedKind, FsFile: fsFile, Blob: gitFile}
		} else {
			return Change{Kind: ModifiedKind, FsFile: fsFile, Blob: gitFile}
		}
	} else if gitFile != nil {
		return Change{Kind: DeletedKind, Blob: gitFile}
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

type Status struct {
	params       Parameters
	fsFiles      []FSFile
	gitBlobs     []git.Blob
	fsFileIndex  map[string]FSFile
	gitFileIndex map[string]git.Blob
	gitStatus    []git.StatusLine
}

func Start(params Parameters) {
	CheckSource(params.Source)
	CheckDestination(params.Destination)
	params.Source = filepath.Clean(params.Source)
	params.Destination = filepath.Clean(params.Destination)
	status := Status{
		params: params,
	}

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

	err = os.Chdir(params.Source)
	if err != nil {
		log.Fatal(err)
	}

	status.fsFiles = FsReadDir(params.Source)
	status.gitBlobs = git.RunGetTreeForHead()
	status.gitStatus = git.RunStatus()

	status.fsFileIndex = Index(status.fsFiles, func(file FSFile) string {
		return file.CleanPath
	})
	status.gitFileIndex = Index(status.gitBlobs, func(file git.Blob) string {
		return file.CleanPath
	})
	var gitIgnoredFiles []string
	for _, status := range status.gitStatus {
		switch v := status.(type) {
		case *git.IgnoredStatusLine:
			gitIgnoredFiles = append(gitIgnoredFiles, v.Path)
		default:
		}
	}
	gitIgnoredFilesIndex := Index(gitIgnoredFiles, func(file string) string {
		return file
	})

	common, left, right := CompareFiles(status.fsFiles, status.gitBlobs, status.fsFileIndex, status.gitFileIndex)

	// Filter out ignored
	//var aLeft []FSFile
	//for _, f := range left {
	//	if _, found := gitIgnoredFilesIndex[f.filepath]; !found {
	//		aLeft = append(aLeft, f)
	//	}
	//}

	var allChanges []Change
	for _, f := range left {
		change := ComparePair(nil, &f, gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}
	for _, f := range right {
		change := ComparePair(&f, nil, gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}
	for _, c := range common {
		change := ComparePair(&c.GitFile, &c.FsFile, gitIgnoredFilesIndex, params.Destination)
		allChanges = append(allChanges, change)
	}

	for _, c := range allChanges {
		switch c.Kind {
		case AddedKind, ModifiedKind, DeletedKind:
			PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == UnchangedKind {
			PrintChange(c)
		}
	}
	for _, c := range allChanges {
		if c.Kind == IgnoredByGitKind {
			PrintChange(c)
		}
	}

	ortoIgnores := 0
	ortoDotGitIgnores := 0
	for _, c := range allChanges {
		if c.Kind == IgnoredByOrtoKind {
			ortoIgnores++
			// TODO: fix this:
			if c.FsFile != nil && !strings.HasPrefix(c.FsFile.CleanPath, ".git/") {
				ortoDotGitIgnores++
			}
		}
	}
	if ortoIgnores == ortoDotGitIgnores && ortoIgnores > 0 {
		println("⛔︎ OrtoIgnored", ".git/**")
	} else {
		for _, c := range allChanges {
			if c.Kind == IgnoredByOrtoKind {
				PrintChange(c)
			}
		}
	}

	println("---")

	for _, c := range allChanges {
		//fmt.Printf("%#v,%#v\n", c.FsFile, c.Blob)
		switch c.Kind {
		case AddedKind, ModifiedKind:
			if c.FsFile != nil { // TODO: why checking this here?
				CopyFile(c.FsFile.CleanPath, c.FsFile.CleanPath, params.Destination)
				PrintCopy(c.FsFile.CleanPath, filepath.Join(params.Destination, c.FsFile.CleanPath))
			}
		case DeletedKind:
			if c.Blob != nil { // TODO: why checking this here?
				// TODO: actually do something with deletions
				PrintDel(c.Blob.CleanPath)
			}
		case UnchangedKind:
			if params.Inclusions.UnchangedFiles {
				CopyFile(c.FsFile.CleanPath, c.FsFile.CleanPath, params.Destination)
				PrintCopy(c.FsFile.CleanPath, filepath.Join(params.Destination, c.FsFile.CleanPath))
			}
		}
	}
}
