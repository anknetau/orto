package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/anknetau/orto/git"
	"github.com/anknetau/orto/orto"
)

func t() {
	//contentFromGit("orto/main.go")
	git.GetRawContent("d46e1c1774be2a4bbad9ae7845a954ae8628018a")
	return
	type Change struct {
		Path        string
		Size        int64
		GitChecksum string
		Checksum    string
	}
	type Record struct {
		Start time.Time `json:"start"`
		Add   []Change  `json:"add"`
		Mod   []Change  `json:"mod"`
		Del   []Change  `json:"del"`
		ID    int       `json:"id"`
	}
	sampleRecord := Record{
		Start: time.Now(),
		ID:    42,
		Add: []Change{
			{
				Path:        "a",
				Size:        math.MaxInt64,
				GitChecksum: "12313",
				Checksum:    "a",
			},
		},
		Mod: []Change{},
		Del: []Change{},
	}
	out, err := json.MarshalIndent(sampleRecord, "", "   ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

	var rec Record
	if err := json.Unmarshal(out, &rec); err != nil {
		panic(err)
	}
}

func main() {
	t()
	os.Exit(0)
	orto.Start(orto.Parameters{
		// "/Users/ank/dev/accounting/accounting"
		// "/Users/ank/dev/mirrors"
		Source:        "/Users/ank/dev/orto",
		Destination:   "/Users/ank/dev/orto/dest/../dest/.", // TODO: this is within the source!
		ChangeSetName: "2025-01-01-my-thing",
		//CopyDotGit:          false,
		//CopyGitIgnoredFiles: false,
		//CopyUnchangedFiles:  false,
	})
}
