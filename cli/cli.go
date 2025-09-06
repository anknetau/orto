package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/anknetau/orto/orto"
	"github.com/anknetau/orto/util"
)

func usage(fs *flag.FlagSet) {
	util.PrintErrf("orto v%s usage:\n", orto.Version())
	util.PrintErrf("orto [flags] <input_dir> <output_dir>\n\n")
	util.PrintErrf("Flags are:\n")
	fs.SetOutput(os.Stderr)
	fs.PrintDefaults()
}

var (
	ErrFatal = errors.New("oh noes")
)

func ParseOrExit() orto.UserParameters {
	result := orto.UserParameters{}
	fs := flag.NewFlagSet("myprog", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	boolFlag := fs.Bool("help", false, "show help")
	fs.StringVar(&result.ChangeSetName, "ChangeSetName", "", "changesetname")

	err := fs.Parse(os.Args[1:])

	if err != nil {
		usageFatal(err)
		os.Exit(2)
	}

	if *boolFlag {
		usage(fs)
		os.Exit(2)
	}

	//fmt.fprintln(, "n =", *nFlag)
	if len(fs.Args()) != 2 {
		usageFatal("Invalid number of arguments. See 'orto --help'")
	}
	println("starting...")
	util.Debug(fs.Args())
	//input := fs.Args()[0]
	//output := fs.Args()[1]
	return result
}

func usageFatal(message ...any) {
	_, _ = fmt.Fprint(os.Stderr, message...)
	_, _ = fmt.Fprint(os.Stderr, "\n")
	os.Exit(2)
}
