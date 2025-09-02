package main

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/anknetau/orto/orto"
)

func jsonOutputTest() {
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
	orto.Start(orto.Parameters{
		// "/Users/ank/dev/accounting/accounting"
		// "/Users/ank/dev/mirrors"
		Source:        "/Users/ank/dev/orto/orto",
		Destination:   "/Users/ank/dev/orto/orto/../dest/.",
		ChangeSetName: "2025-01-01-my-thing",
		//CopyDotGit:          false,
		//CopyGitIgnoredFiles: false,
		//CopyUnchangedFiles:  false,
	})
}
