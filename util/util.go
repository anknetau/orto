package util

import (
	"fmt"
	"os"
)

func Debug(value any) {
	fmt.Printf("%#v\n", value)
	//b, err := json.MarshalIndent(value, "", " ")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//println(string(b))
}

func PrintErrf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
