package fp_test

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/anknetau/orto/assert"
	"github.com/anknetau/orto/fp"
)

func diff(filepath1, filepath2 string) string {
	if !filepath.IsAbs(filepath1) || !filepath.IsAbs(filepath2) {
		log.Fatal("this needs absolute paths!")
	}
	filepath1 = filepath.Clean(filepath1)
	filepath2 = filepath.Clean(filepath2)
	//a := removeSeparators(fp.SplitFilePath(filepath1))
	//b := removeSeparators(fp.SplitFilePath(filepath2))
	return ""
}

func TestDirectoryIsParentOfOrEqual(t *testing.T) {
	assert.Equal(t, true, fp.AbsolutePathIsParentOrEqual("/a", "/a/b"))
	assert.Equal(t, []string{"a"}, fp.FilepathParts("/a"))

	assert.Equal(t, true, fp.AbsolutePathIsParentOrEqual("/a//b/c/d", "/a/b/c/d"))
	assert.Equal(t, false, fp.AbsolutePathIsParentOrEqual("/a", "/b"))

	assert.Equal(t, true, fp.AbsolutePathIsParentOrEqual("/Users/ank/dev/orto/orto", "/Users/ank/dev/orto/orto/dest/../dest/."))
	assert.Equal(t, false, fp.AbsolutePathIsParentOrEqual("/Users/ank/dev/orto/orto/dest/../dest/.", "/Users/ank/dev/orto/orto"))
}

func TestFilepathParts(t *testing.T) {
	assert.Equal(t, []string{"a"}, fp.FilepathParts("/a"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, fp.FilepathParts("/a//b/c/d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, fp.FilepathParts("/a/b/c/d"))
	assert.Equal(t, []string{}, fp.FilepathParts("/a/.."))
	assert.Equal(t, []string{"a", "b"}, fp.FilepathParts("/a/./b"))
	assert.Equal(t, []string{"a"}, fp.FilepathParts("/a///./"))
	assert.Equal(t, []string{"a", "b", "c", "d", "eff123"}, fp.FilepathParts("/a/b/c/d////eff123"))

	assert.Equal(t, []string{}, fp.FilepathParts("/"))

	assert.Equal(t, ".", filepath.Clean(""))
	assert.Equal(t, []string{"."}, fp.FilepathParts(""))
	assert.Equal(t, []string{"."}, fp.FilepathParts("."))
	assert.Equal(t, []string{"."}, fp.FilepathParts("././"))
	assert.Equal(t, []string{"."}, fp.FilepathParts("./././///"))
}

func TestJoined(t *testing.T) {
	abs1 := "/a"
	abs2 := "/a/b"
	_ = diff(abs1, abs2)
	//assert.Equal(t, "", filepath)
}

func TestValidFilePathForOrto(t *testing.T) {
	test := func(input string, valid bool) {
		t.Helper()
		assert.Equal(t, valid, fp.ValidFilePathForOrto(input))
	}
	test("", false)
	test(".", true)
	test("/", true)
	test("./a", true)
	test("../a", true)
	test("../aaaa a", false)
	test(" ", false)
	test("./aaaaa/.", true)
}

func TestCleanFilePath(t *testing.T) {
	test := func(input string) {
		t.Helper()
		assert.Equal(t, filepath.Clean(input), fp.CleanFilePath(input))
	}
	test("")
	test(".")
	test("/")
	test("//")
	test("//a")
	test("c://a")
	test(".//a")
	test("/////a")
	test("aaaa")
	test("aaaa//////")
	test("/////////aaaa//////")
	test(".///.//.//.//aaaa/.//.//./.")
	test(".///.//.//.//aaaa/.//.//./.")
	test("./.")
	test("././")
	test("/././")
	test("a/././")
	test("a/././a")
	test("a//a")
	test("a//a/..")
	test("../../..")
	test("../a/../..")
	test("../a/../")
	test("./a/../")
	test("./a/../////")
	test("../a//a")
	test("../a//a/..")
}

func TestSplitFilePath(t *testing.T) {
	testJoin := func(input string, expected []string) {
		t.Helper()
		assert.Equal(t, expected, fp.SplitFilePath(input))
	}
	testJoin("", []string{})
	testJoin("c:/windows", []string{"c:", "/", "windows"})
	testJoin(".git", []string{".git"})
	testJoin(".git/a", []string{".git", "/", "a"})
	testJoin("/", []string{"/"})
	testJoin("//", []string{"//"})
	testJoin("../", []string{"..", "/"})
	testJoin("a", []string{"a"})
	testJoin("/a", []string{"/", "a"})
	testJoin("////a", []string{"////", "a"})
	testJoin("a/", []string{"a", "/"})
	testJoin("/a/", []string{"/", "a", "/"})
	testJoin("a//a", []string{"a", "//", "a"})
	testJoin("a//b", []string{"a", "//", "b"})
	testJoin("aaaaa", []string{"aaaaa"})
	testJoin("./aaaa/blah", []string{".", "/", "aaaa", "/", "blah"})
	testJoin("./aaaa/blah/", []string{".", "/", "aaaa", "/", "blah", "/"})
	testJoin("/////", []string{"/////"})
	testJoin("a/////a////", []string{"a", "/////", "a", "////"})
	testJoin("/////a////a", []string{"/////", "a", "////", "a"})
	testJoin("/////a////a", []string{"/////", "a", "////", "a"})
}

func TestSplitFilePath_NoEmptyElements(t *testing.T) {
	cases := []string{
		"a//b",
		"///",
		"/a//b//",
		"a////",
		"",
		"a",
		"/a/b/c",
	}
	for _, s := range cases {
		parts := fp.SplitFilePath(s)
		for _, p := range parts {
			assert.NotEqual(t, "", p)
		}
	}
}

func TestSplitFilePath_ExpectedShapes(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"a//b", []string{"a", "//", "b"}},
		{"///", []string{"///"}},
		{"/a//b//", []string{"/", "a", "//", "b", "//"}},
		{"a////", []string{"a", "////"}},
		{"", []string{}},
		{"a", []string{"a"}},
	}
	for _, tt := range tests {
		got := fp.SplitFilePath(tt.in)
		assert.Equal(t, tt.want, got)
	}
}
