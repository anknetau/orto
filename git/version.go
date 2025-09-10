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

func runToString(pathToBinary string, args ...string) (string, error) {
	cmd := exec.Command(pathToBinary, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func RunVersion(pathToBinary string) string {
	output, err := runToString(pathToBinary, "--version")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return ""
		} else {
			log.Fatalf("Cannot execute git: %s", err)
		}
	}

	output = strings.TrimSpace(output)
	// Version looks like major.minor.path(-rcN)
	version, ok := strings.CutPrefix(output, "git version ")
	if !ok {
		log.Fatalf("Version response from git not recognized: [%s]", output)
	}
	matches := reGitVersion.FindString(version)
	if len(matches) == 0 {
		log.Fatalf("Version response from git not recognized: [%s]", output)
	}
	return version
}
