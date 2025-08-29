package main

import (
	"github.com/anknetau/orto/orto"
)

func main() {
	orto.Start(orto.Parameters{
		// "/Users/ank/dev/accounting/accounting"
		// "/Users/ank/dev/mirrors"
		Source:              "/Users/ank/dev/orto",
		Destination:         "/Users/ank/dev/orto/dest/../dest/.", // TODO: this is within the source!
		CopyDotGit:          false,
		CopyGitIgnoredFiles: false,
		CopyUnchangedFiles:  true,
	})
}
