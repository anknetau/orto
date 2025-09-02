package git_test

import (
	"testing"

	"github.com/anknetau/orto/assert"
	"github.com/anknetau/orto/git"
)

// TODO: test all sorts of weird characters/utf8/spaces/etc

func assertLine[T any](t *testing.T, expected T, line string) {
	t.Helper()
	assert.Equal(t, expected, git.ParseLine(line).(T))
}

func TestSamples(t *testing.T) {
	// Fake status lines:
	assertLine(t, git.ChangedStatusLine{
		Xy:   ".M",
		Sub:  "N...",
		MH:   "100644",
		MI:   "000000",
		MW:   "100755",
		HH:   "21809e0abf6af128398a1687adf8a0fc22d1ca88",
		HI:   "45b983be36b73c0788dc9cbcb76cbb80fc7bb057",
		Path: ".idea/dictionaries/project.xml"},
		"1 .M N... 100644 000000 100755 21809e0abf6af128398a1687adf8a0fc22d1ca88 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 .idea/dictionaries/project.xml")
	// Real status lines:
	assertLine(t, git.ChangedStatusLine{
		Xy:   ".M",
		Sub:  "N...",
		MH:   "100644",
		MI:   "100644",
		MW:   "100644",
		HH:   "21809e0abf6af128398a1687adf8a0fc22d1ca88",
		HI:   "21809e0abf6af128398a1687adf8a0fc22d1ca88",
		Path: ".idea/dictionaries/project.xml"},
		"1 .M N... 100644 100644 100644 21809e0abf6af128398a1687adf8a0fc22d1ca88 21809e0abf6af128398a1687adf8a0fc22d1ca88 .idea/dictionaries/project.xml")
	assertLine(t, git.RenamedOrCopiedStatusLine{Change: git.ChangedStatusLine{
		Xy:   "R.",
		Sub:  "N...",
		MH:   "100644",
		MI:   "100644",
		MW:   "100644",
		HH:   "45b983be36b73c0788dc9cbcb76cbb80fc7bb057",
		HI:   "45b983be36b73c0788dc9cbcb76cbb80fc7bb057",
		Path: "deleteme2"}, OrigPath: "deleteme", Score: "R100"},
		"2 R. N... 100644 100644 100644 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 R100 deleteme2\x00deleteme")
	assertLine(t, git.UntrackedStatusLine{Path: "gitdiff.txt"}, "? gitdiff.txt")
	assertLine(t, git.IgnoredStatusLine{Path: "gitdiff.txt"}, "! gitdiff.txt")
	//# branch.oid 35539293fc213ca0e573d35cae496b56a0f4ab06
	//# branch.head master
	//# branch.upstream origin/master
	//# branch.ab +0 -0
	//1 .M N... 100644 100644 100644 21809e0abf6af128398a1687adf8a0fc22d1ca88 21809e0abf6af128398a1687adf8a0fc22d1ca88 .idea/dictionaries/project.xml
	//2 R. N... 100644 100644 100644 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 R100 deleteme2
	//deleteme
	//1 .M N... 100644 100644 100644 37ee65c344d8ab16aebbed88699b77f3a0f2ee7f 37ee65c344d8ab16aebbed88699b77f3a0f2ee7f git/blob.go
	//	1 .M N... 100644 100644 100644 a1b4254081e1cc3a404e4fc7a5282ec1562a2585 a1b4254081e1cc3a404e4fc7a5282ec1562a2585 git/status.go
	//	1 AM N... 000000 100644 100644 0000000000000000000000000000000000000000 e69de29bb2d1d6434b8b29ae775ad8c2e48c5391 git/status_test.go
	//? diff.txt
	//? diff2.txt
	//? gitdiff.txt
	//! .idea/workspace.xml
}

func TestComments(t *testing.T) {
	lines := "# branch.oid 35539293fc213ca0e573d35cae496b56a0f4ab06\x00# branch.head master\x00# branch.upstream origin/master\x00# branch.ab +0 -0"
	statusLines := git.ParseLines(lines)
	if newLine, ok := statusLines[0].(git.CommentStatusLine); ok {
		// do something with newLine
	}
}
