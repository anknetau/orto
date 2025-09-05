package git

import (
	"log"
	"os/exec"
	"strings"

	"github.com/anknetau/orto/fp"
)

func RunGetTreeForHead(config fp.EnvConfig) ([]Blob, []Submodule) {
	// %(objectmode) %(objecttype) %(objectname)%x09%(path)
	cmd := exec.Command(config.GitCommand, "ls-tree", "HEAD", "-r", "--format=%(objecttype)|>%(objectname)|>%(path)|>%(objectmode)", "-z")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var blobs []Blob
	var submodules []Submodule
	output := strings.TrimRight(string(out), "\x00")
	lines := strings.SplitSeq(output, "\x00")

	for line := range lines {
		pBlob, pSubmodule := parseGetTreeLine(line)
		if pBlob == nil && pSubmodule == nil {
			log.Fatal("Invalid line from git: " + line)
		}
		if pBlob != nil {
			blobs = append(blobs, *pBlob)
		} else {
			submodules = append(submodules, *pSubmodule)
		}
	}

	return blobs, submodules
}
