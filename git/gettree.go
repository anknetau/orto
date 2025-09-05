package git

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/anknetau/orto/fp"
)

func RunGetTreeForHead(config fp.EnvConfig) []Blob {
	// %(objectmode) %(objecttype) %(objectname)%x09%(path)
	cmd := exec.Command(config.GitCommand, "ls-tree", "HEAD", "-r", "--format=%(objecttype)|>%(objectname)|>%(path)|>%(objectmode)", "-z")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var result []Blob
	output := strings.TrimRight(string(out), "\x00")
	lines := strings.SplitSeq(output, "\x00")

	for line := range lines {
		gitFile := gitFileFromLine(line)
		result = append(result, gitFile)
	}

	return result
}

func gitFileFromLine(line string) Blob {
	fields := strings.Split(line, "|>")
	fmt.Printf("fields: %#v\n", fields)
	if len(fields) != 4 || len(fields[1]) == 0 || len(fields[2]) == 0 || len(fields[3]) == 0 {
		log.Fatal("Invalid line from git: " + line)
	}
	objectType := fields[0]
	path := fields[2]
	checksum := fp.NewChecksum(fields[1])
	mode := NewMode(fields[3])
	return NewGitFile(objectType, path, checksum, mode)
}
