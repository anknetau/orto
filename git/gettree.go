package git

import (
	"log"
	"os/exec"
	"strings"

	"github.com/anknetau/orto/fp"
)

func RunGetTreeForHead(config fp.EnvConfig) []Blob {
	cmd := exec.Command(config.GitCommand, "ls-tree", "HEAD", "-r", "--format=%(objecttype)|>%(objectname)|>%(path)|>%(objectmode)")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var result []Blob
	output := string(out)
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")

	for line := range lines {
		gitFile := gitFileFromLine(line)
		result = append(result, gitFile)
	}

	return result
}

func gitFileFromLine(line string) Blob {
	fields := strings.Split(line, "|>")
	if len(fields) != 4 || len(fields[1]) == 0 || len(fields[2]) == 0 || len(fields[3]) == 0 {
		log.Fatal("Invalid line from git: " + line)
	}
	objectType := fields[0]
	path := fields[2]
	checksum := fp.NewChecksum(fields[1])
	mode := NewMode(fields[3])
	return NewGitFile(objectType, path, checksum, mode)
}
