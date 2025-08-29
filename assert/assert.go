package assert

import (
	"fmt"
	"reflect"
	"testing"
)

func True(t *testing.T, value bool, message ...interface{}) {
	t.Helper()
	if !value {
		Fail(t, firstAsString(message))
	}
}

func False(t *testing.T, value bool, message ...interface{}) {
	t.Helper()
	if value {
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

func Equal[T any](t *testing.T, expected T, actual T, message ...interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		s := ""
		userMsg := firstAsString(message...)
		if len(userMsg) > 0 {
			s = fmt.Sprintf(" (%s)", userMsg)
		}
		Fail(t, fmt.Sprintf("Expected: '%#v'. Actual: '%#v'%s", expected, actual, s))
	}
}

func NotEqual[T any](t *testing.T, expected T, actual T, message ...interface{}) {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		s := ""
		userMsg := firstAsString(message...)
		if len(userMsg) > 0 {
			s = fmt.Sprintf(" (%s)", userMsg)
		}
		Fail(t, fmt.Sprintf("Expected To Not Be: '%#v'. Actual: '%#v'%s", expected, actual, s))
	}
}

func firstAsString(arg ...interface{}) string {
	if len(arg) == 0 {
		return ""
	}
	if s, ok := arg[0].(string); ok {
		return s
	}
	return ""
}
