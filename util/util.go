package util

import (
	"fmt"
	"os"
	"time"
)

func Debug(value any) {
	fmt.Printf("%#v\n", value)
	//b, err := json.MarshalIndent(value, "", " ")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//println(string(b))
}

func ErrPrintLnf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	_, _ = fmt.Fprintf(os.Stderr, "\n")
}

// SerializedDateTime returns a string that looks like "2006-01-02_15-04-05"
func SerializedDateTime(now time.Time) string {
	return now.Format("2006-01-02_15-04-05")
}
