package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/anknetau/orto/orto"
	"github.com/anknetau/orto/util"
)

func printUsage(fs *flag.FlagSet) {
	util.ErrPrintLnf("orto v%s usage:\n", orto.Version())
	util.ErrPrintLnf("orto [flags] <input_dir> <output_dir>\n")
	util.ErrPrintLnf("Flags are:\n")
	util.ErrPrintLnf("  -help: show help")
	fs.SetOutput(os.Stderr)
	fs.PrintDefaults()
	flag.PrintDefaults()
	util.ErrPrintLnf("")
	util.ErrPrintLnf("Flags can be passed with one or two dashes: -x and --x are equivalent\n")
}

var (
	ErrFatal = errors.New("oh noes")
)

func ParseOrExit() orto.UserParameters {
	result := orto.UserParameters{}
	flagSet := flag.NewFlagSet("orto", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.Usage = func() {}
	//boolFlag := flagSet.Bool("help", false, "show help")
	now := util.SerializedDateTime(time.Now())
	flagSet.StringVar(&result.ChangeSetName, "ChangeSetName", "", "ChangeSetName to use. Default: current datetime (eg '"+now+"')")

	err := flagSet.Parse(os.Args[1:])

	if errors.Is(err, flag.ErrHelp) {
		printUsage(flagSet)
		os.Exit(2)
	}
	if err != nil {
		usageFatal(err)
		os.Exit(2)
	}

	//fmt.fprintln(, "n =", *nFlag)
	if len(flagSet.Args()) != 2 {
		usageFatal("Invalid number of arguments. See 'orto -h'")
	}
	println("starting...")
	util.Debug(flagSet.Args())
	//input := flagSet.Args()[0]
	//output := flagSet.Args()[1]
	return result
}

func usageFatal(message ...any) {
	_, _ = fmt.Fprint(os.Stderr, message...)
	_, _ = fmt.Fprint(os.Stderr, "\n")
	os.Exit(2)
}
