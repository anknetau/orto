package orto

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func GitRunGetTreeForHead() []GitFile {
	cmd := exec.Command("git", "ls-tree", "HEAD", "-r", "--format=%(objecttype)|>%(objectname)|>%(path)|>%(objectmode)")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var result []GitFile
	output := string(out)
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")

	for line := range lines {
		fields := strings.Split(line, "|>")
		if len(fields) != 4 || len(fields[1]) == 0 || len(fields[2]) == 0 || len(fields[3]) == 0 {
			log.Fatal("Invalid line from git: " + line)
		}
		// When `git ls-tree` is passed -r, it will recurse and not show trees, but resolve the blobs within instead.
		// TODO: Submodules will appear as a `commit`
		if fields[0] != "blob" {
			log.Fatal("Unsupported object type from git: " + line)
		}
		gitFile := GitFile{
			Filepath: filepath.Clean(fields[2]),
			path:     fields[2],
			checksum: fields[1],
			mode:     fields[3],
		}
		if checksumGetAlgo(gitFile.checksum) == UNKNOWN {
			log.Fatal("Don't know how this was hashed: " + line)
		}
		result = append(result, gitFile)
	}

	return result
}
