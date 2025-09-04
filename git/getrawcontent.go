package git

import (
	"log"
	"os/exec"

	"github.com/anknetau/orto/fp"
)

func RunGetRawContent(checksum fp.Checksum) []byte {
	// TODO: stream this rather than load it all into memory
	cmd := exec.Command("git", "cat-file", "blob", string(checksum))
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return out
}

//let's stage this
