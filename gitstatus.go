package main

import (
	"log"
	"os/exec"
	"strings"
)

func gitRunStatus() []StatusLine {
	cmd := exec.Command("git", "status", "--porcelain=v2", "--untracked-files=all", "--show-stash", "--branch", "--ignored")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	var result []StatusLine
	output := string(out)
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")
	for line := range lines {
		if strings.HasPrefix(line, "#") {
			// TODO: comment
			continue
		}
		if path, found := strings.CutPrefix(line, "! "); found {
			f := IgnoredStatusLine{path: path}
			result = append(result, &f)
		} else {
			println(">", line)
		}
	}
	return result
}
