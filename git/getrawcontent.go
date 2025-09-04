package git

import (
	"log"
	"os/exec"
)

func RunGetRawContent(checksum string) []byte {
	// TODO: stream this rather than load it all into memory
	cmd := exec.Command("git", "cat-file", "blob", checksum)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return out
}

//let's stage this
