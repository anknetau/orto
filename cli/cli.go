package cli

import (
	"flag"

	"github.com/anknetau/orto/orto"
)

func CliParse() orto.UserParameters {
	var nFlag = flag.Int("n", 1234, "help message for flag n")
	flag.Parse()
	println(*nFlag)
	return orto.UserParameters{}
}
