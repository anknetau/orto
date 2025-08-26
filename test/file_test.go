package orto_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/anknetau/orto/orto"
)

func TestSpliterino(t *testing.T) {
	//t.Error(filepath.Clean("/////"))
	testJoin := func(input string, expected string) string {
		p := strings.ReplaceAll(input, string(filepath.Separator), "/")
		result := strings.Join(orto.SplitFilePath(p), ",")
		if result != expected {
			t.Errorf("SplitFilePath(%q): got %q, want %q", p, result, expected)
		}
		return result
	}
	testJoin("", "")
	testJoin("c:/windows", "c:,/,windows")
	testJoin(".git", ".git")
	testJoin(".git/a", ".git,/,a")
	testJoin("/", "/")
	testJoin("//", "//")
	testJoin("../", "..,/")
	testJoin("a", "a")
	testJoin("/a", "/,a")
	testJoin("////a", "////,a")
	testJoin("a/", "a,/")
	testJoin("/a/", "/,a,/")
	testJoin("a//a", "a,//,a")
	testJoin("aaaaa", "aaaaa")
	testJoin("./aaaa/blah", ".,/,aaaa,/,blah")
	testJoin("./aaaa/blah/", ".,/,aaaa,/,blah,/")
	testJoin("/////", "/////")
	testJoin("a/////a////", "a,/////,a,////")
	testJoin("/////a////a", "/////,a,////,a")
}

//func TestValidFilePath(t *testing.T) {
//	valid := []string{
//		"a",
//		"/",
//		"./",
//		"../",
//		"./a",
//		"../a",
//		"/a",
//		"a/",
//		"/a/",
//	}
//	invalid := []string{
//		"",
//		"//",
//		"///",
//		"a///",
//		"///a",
//		"//a",
//	}
//	for _, c := range valid {
//		if orto.ValidFilePath(c) != true {
//			t.Errorf("Should be valid: %s", c)
//		}
//	}
//	for _, c := range invalid {
//		if orto.ValidFilePath(c) != false {
//			t.Errorf("Should be invalid: %s", c)
//		}
//	}
//}
