package git

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var (
	reGitVersion = regexp.MustCompile(`^(\d+)[.](\d+)[.](\d+(-rc\d+)?)$`)
)

func RunVersion(gitCommand string) *string {
	cmd := exec.Command(gitCommand, "--version")
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil
		} else {
			log.Fatalf("Cannot execute git: %s", err)
		}
	}
	output := string(out)
	output = strings.TrimSpace(output)
	version, ok := strings.CutPrefix(output, "git version ")
	if !ok {
		log.Fatalf("Version response from git not recognized: [%s]", output)
	}
	matches := reGitVersion.FindString(version)
	if len(matches) == 0 {
		log.Fatalf("Version response from git not recognized: [%s]", output)
	}
	return &version
}
