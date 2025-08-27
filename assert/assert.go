package assert

import (
	"fmt"
	"testing"
)

func True(t *testing.T, value bool, message ...interface{}) {
	t.Helper()
	if !value {
		Fail(t, firstAsString(message))
	}
}

func Fail(t *testing.T, message string) {
	t.Helper()
	if len(message) == 0 {
		message = "Assertion failed"
	}
	t.Errorf("%s: %s", t.Name(), message)
}

func Equal(t *testing.T, expected string, actual string, message ...interface{}) {
	t.Helper()
	if expected != actual {
		s := ""
		userMsg := firstAsString(message...)
		if len(userMsg) > 0 {
			s = fmt.Sprintf(" (%s)", userMsg)
		}
		Fail(t, fmt.Sprintf("Expected: '%s'. Actual: '%s'%s", expected, actual, s))
	}
}

func firstAsString(arg ...interface{}) string {
	if arg == nil || len(arg) < 1 {
		return ""
	}
	if s, ok := arg[0].(string); ok {
		return s
	}
	return ""
}
