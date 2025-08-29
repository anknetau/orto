package fp_test

import (
	"path/filepath"
	"testing"

	"github.com/anknetau/orto/assert"
	"github.com/anknetau/orto/fp"
)

func TestJoined(t *testing.T) {
	assert.Equal(t, "/a/c", filepath.Join("/", "a", "c"))
	assert.Equal(t, "/a/c", filepath.Join("/a/.", "./c"))
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
