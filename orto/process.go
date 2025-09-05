package orto

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/anknetau/orto/fp"
	"github.com/anknetau/orto/git"
)

type GitBlobAndFile struct {
	FsFile  FSFile
	GitBlob git.Blob
}

func Filter[T any, U any](items []T, callback func(*T) *U) []U {
	var result = make([]U, 0, len(items))
	for _, item := range items {
		val := callback(&item)
		if val != nil {
			result = append(result, *val)
		}
	}
	return result
}

func Index[T any](items []T, callback func(T) string) map[string]T {
	fsFileIndex := make(map[string]T, len(items))
	for _, fsFile := range items {
		fsFileIndex[callback(fsFile)] = fsFile
	}
	return fsFileIndex
}

func copyContents(src *os.File, dest *os.File) int64 {
	n, err := dest.ReadFrom(src)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func SaveGitBlob(gitCommand string, checksum fp.Checksum, path string, destAbsoluteDirectory string) {
	fp.CreateIntermediateDirectoriesForFile(path, destAbsoluteDirectory)

	content := git.RunGetRawContent(gitCommand, checksum)
	err := os.WriteFile(filepath.Join(destAbsoluteDirectory, path), content, 0644)
	if err != nil {
		log.Fatal(err)
	}
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

	// TODO: we are assuming here that this is a file and not a directory.
	fp.CreateIntermediateDirectoriesForFile(destRelativePath, destAbsoluteDirectory)

	read, err := os.Open(sourceRelativePath)
	if err != nil {
		log.Fatal(err)
	}
	defer read.Close()

	write, err := os.Create(destAbsoluteFile)
	if err != nil {
		log.Fatal(err)
	}
	defer write.Close()
	return copyContents(read, write)
}

func PrintLogHeader(s string) {
	println("✴️ " + s)
}

func PrintLogCopy(src string, dst string) {
	println("  🔹" + src + " → " + dst)
}

func PrintLogDel(src string) {
	println("  🔹" + src + " ❌ ")
}

func PrintChange(change Change) {
	switch change.Kind {
	case ChangeKindAdded:
		println("  ❇️ Added", change.FsFile.CleanPath)
	case ChangeKindDeleted:
		println("  ❌ Deleted", change.GitBlob.CleanPath)
	case ChangeKindUnchanged:
		println("  ➖ Unchanged", change.FsFile.CleanPath)
	case ChangeKindModified:
		println("  ✏️ Modified", change.FsFile.CleanPath)
	case ChangeKindIgnoredByGit:
		println("  ⛔︎ GitIgnored", change.FsFile.CleanPath)
	case ChangeKindIgnoredByOrto:
		println("  ⛔︎ OrtoIgnored", change.FsFile.CleanPath)
	}
}

func CompareFiles(gitBlobs []git.Blob, fsFiles []FSFile, fsFileIndex map[string]FSFile, gitBlobIndex map[string]git.Blob) ([]GitBlobAndFile, []FSFile, []git.Blob) {
	var common []GitBlobAndFile
	var resultFsFiles []FSFile
	var resultGitBlobs []git.Blob
	for _, fsFile := range fsFiles {
		gitBlob, ok := gitBlobIndex[fsFile.CleanPath]
		if ok {
			common = append(common, GitBlobAndFile{FsFile: fsFile, GitBlob: gitBlob})
		} else {
			resultFsFiles = append(resultFsFiles, fsFile)
		}
	}

	for _, gitBlob := range gitBlobs {
		_, ok := fsFileIndex[gitBlob.CleanPath]
		if !ok {
			resultGitBlobs = append(resultGitBlobs, gitBlob)
		}
	}
	return common, resultFsFiles, resultGitBlobs
}

func isOrtoIgnored(fsFile *FSFile, inputSettings InputSettings) bool {
	splitParts := fp.SplitFilePath(fsFile.CleanPath)
	if !inputSettings.copyDotGit && len(splitParts) > 0 && splitParts[0] == ".git" {
		return true
	}
	if len(splitParts) > 0 && splitParts[0] == "third_party" || splitParts[0] == "laf" {
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

func ComparePair(gitBlob *git.Blob, fsFile *FSFile, gitIgnoredFilesIndex map[string]string, inputSettings InputSettings) Change {
	if fsFile != nil {
		if isOrtoIgnored(fsFile, inputSettings) {
			return Change{Kind: ChangeKindIgnoredByOrto, FsFile: fsFile}
		}
		if _, ignored := gitIgnoredFilesIndex[fsFile.CleanPath]; ignored {
			return Change{Kind: ChangeKindIgnoredByGit, FsFile: fsFile}
		}
	}
	if gitBlob == nil && fsFile == nil {
		panic("Illegal state")
	}
	if gitBlob != nil && fsFile != nil {
		if gitBlob.CleanPath != fsFile.CleanPath {
			panic("Illegal state: " + gitBlob.CleanPath + " " + fsFile.CleanPath)
		}
		// TODO: os.Stat follows symlinks apparently
		stat, err := os.Stat(fsFile.Path)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			panic("was dir: " + fsFile.Path)
		}
		calculatedChecksum := fp.ChecksumBlob(gitBlob.Path, gitBlob.Checksum.GetAlgo())
		if calculatedChecksum == gitBlob.Checksum {
			return Change{Kind: ChangeKindUnchanged, FsFile: fsFile, GitBlob: gitBlob}
		} else {
			return Change{Kind: ChangeKindModified, FsFile: fsFile, GitBlob: gitBlob}
		}
	} else if gitBlob != nil {
		return Change{Kind: ChangeKindDeleted, GitBlob: gitBlob}
	} else {
		return Change{Kind: ChangeKindAdded, FsFile: fsFile}
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
