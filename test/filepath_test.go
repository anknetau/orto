package orto_test

import (
	"path/filepath"
	"testing"

	"github.com/anknetau/orto/assert"
	"github.com/anknetau/orto/orto"
)

func TestCleanFilePath(t *testing.T) {
	test := func(input string) {
		assert.Equal(t, filepath.Clean(input), orto.CleanFilePath(input))
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
		result := orto.SplitFilePath(input)
		assert.Equal(t, expected, result)
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
	testJoin("aaaaa", []string{"aaaaa"})
	testJoin("./aaaa/blah", []string{".", "/", "aaaa", "/", "blah"})
	testJoin("./aaaa/blah/", []string{".", "/", "aaaa", "/", "blah", "/"})
	testJoin("/////", []string{"/////"})
	testJoin("a/////a////", []string{"a", "/////", "a", "////"})
	testJoin("/////a////a", []string{"/////", "a", "////", "a"})
}
