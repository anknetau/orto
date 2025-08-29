package orto

import (
	"log"
	"os/exec"
	"strings"
)

func GitRunGetTreeForHead() []GitFile {
	//goland:noinspection SpellCheckingInspection
	cmd := exec.Command("git", "ls-tree", "HEAD", "-r", "--format=%(objecttype)|>%(objectname)|>%(path)|>%(objectmode)")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var result []GitFile
	output := string(out)
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")

	for line := range lines {
		gitFile := gitFileFromLine(line)
		result = append(result, gitFile)
	}

	return result
}

func gitFileFromLine(line string) GitFile {
	fields := strings.Split(line, "|>")
	if len(fields) != 4 || len(fields[1]) == 0 || len(fields[2]) == 0 || len(fields[3]) == 0 {
		log.Fatal("Invalid line from git: " + line)
	}
	objectType := fields[0]
	path := fields[2]
	checksum := fields[1]
	mode := fields[3]
	return MakeGitFile(objectType, path, checksum, mode)

}
