package main

import (
	"github.com/anknetau/orto/orto"
)

func main() {
	params := orto.Parameters{
		// "/Users/ank/dev/accounting/accounting"
		// "/Users/ank/dev/mirrors"
		Source:      "/Users/ank/dev/orto",
		Destination: "/Users/ank/dev/orto/dest", // TODO: this is within the source!
	}
	orto.Start(params)
}
