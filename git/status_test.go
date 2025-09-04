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
		Status:        ".M",
		Sub:           "N...",
		ModeHead:      "100644",
		ModeIndex:     "000000",
		ModeWorktree:  "100755",
		ChecksumHead:  "21809e0abf6af128398a1687adf8a0fc22d1ca88",
		ChecksumIndex: "45b983be36b73c0788dc9cbcb76cbb80fc7bb057",
		Path:          ".idea/dictionaries/project.xml"},
		"1 .M N... 100644 000000 100755 21809e0abf6af128398a1687adf8a0fc22d1ca88 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 .idea/dictionaries/project.xml")
	// Real status lines:
	assertLine(t, git.ChangedStatusLine{
		Status:        ".M",
		Sub:           "N...",
		ModeHead:      "100644",
		ModeIndex:     "100644",
		ModeWorktree:  "100644",
		ChecksumHead:  "21809e0abf6af128398a1687adf8a0fc22d1ca88",
		ChecksumIndex: "21809e0abf6af128398a1687adf8a0fc22d1ca88",
		Path:          ".idea/dictionaries/project.xml"},
		"1 .M N... 100644 100644 100644 21809e0abf6af128398a1687adf8a0fc22d1ca88 21809e0abf6af128398a1687adf8a0fc22d1ca88 .idea/dictionaries/project.xml")
	assertLine(t, git.RenamedOrCopiedStatusLine{Change: git.ChangedStatusLine{
		Status:        "R.",
		Sub:           "N...",
		ModeHead:      "100644",
		ModeIndex:     "100644",
		ModeWorktree:  "100644",
		ChecksumHead:  "45b983be36b73c0788dc9cbcb76cbb80fc7bb057",
		ChecksumIndex: "45b983be36b73c0788dc9cbcb76cbb80fc7bb057",
		Path:          "deleteme2"}, OrigPath: "deleteme", Score: "R100"},
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

	// More real examples:
	//StatusLineKindComment: Comment: branch.oid 581953ffc6256bd3a33d3d459e592f30949ff6c9
	//StatusLineKindComment: Comment: branch.head master
	//StatusLineKindComment: Comment: branch.upstream origin/master
	//StatusLineKindComment: Comment: branch.ab +0 -0
	//StatusLineKindChanged: {Path:assets/orto-git.png Status:.D Sub:N... ModeHead:100644 ModeIndex:100644 ModeWorktree:000000 ChecksumHead:2abf629e87264b60acb577dc7efebdd618b2bfa4 ChecksumIndex:2abf629e87264b60acb577dc7efebdd618b2bfa4}
	//StatusLineKindChanged: {Path:deleteme Status:.D Sub:N... ModeHead:100644 ModeIndex:100644 ModeWorktree:000000 ChecksumHead:45b983be36b73c0788dc9cbcb76cbb80fc7bb057 ChecksumIndex:45b983be36b73c0788dc9cbcb76cbb80fc7bb057}
	//StatusLineKindChanged: {Path:fp/fp.go Status:.M Sub:N... ModeHead:100644 ModeIndex:100644 ModeWorktree:100644 ChecksumHead:6474594b176af87916439279160a1f26964d00ec ChecksumIndex:6474594b176af87916439279160a1f26964d00ec}
	//StatusLineKindChanged: {Path:git/version.go Status:A. Sub:N... ModeHead:000000 ModeIndex:100644 ModeWorktree:100644 ChecksumHead:0000000000000000000000000000000000000000 ChecksumIndex:4607879d88f1d95671d9664bf6dde92ed23d43ea}
	//StatusLineKindChanged: {Path:orto/main.go Status:M. Sub:N... ModeHead:100644 ModeIndex:100644 ModeWorktree:100644 ChecksumHead:4456164e5e61db3dc1f201421c53284366af6529 ChecksumIndex:9e93db03222710334023edf0cbd1527e62667eca}
	//StatusLineKindChanged: {Path:orto/params.go Status:M. Sub:N... ModeHead:100644 ModeIndex:100644 ModeWorktree:100644 ChecksumHead:ba9d176ee5149d3a4879ed6ac19aadcaf29fe903 ChecksumIndex:1d12c1aa8a718dd74dceb352e3b07b4375b7eb57}
	//StatusLineKindUntracked: Path:deleteme2
	//StatusLineKindUntracked: Path:diff.txt
	//StatusLineKindUntracked: Path:diff2.txt
	//StatusLineKindUntracked: Path:gitdiff.txt
	//StatusLineKindIgnored: Path:.idea/workspace.xml
}

func TestComments(t *testing.T) {
	lines := "# branch.oid 35539293fc213ca0e573d35cae496b56a0f4ab06\x00# branch.head master\x00# branch.upstream origin/master\x00# branch.ab +0 -0"
	statusLines := git.ParseLines(lines)
	assert.Equal(t, 4, len(statusLines))
	assert.Equal(t, " branch.oid 35539293fc213ca0e573d35cae496b56a0f4ab06", statusLines[0].(git.CommentStatusLine).Comment)
	assert.Equal(t, " branch.head master", statusLines[1].(git.CommentStatusLine).Comment)
	assert.Equal(t, " branch.upstream origin/master", statusLines[2].(git.CommentStatusLine).Comment)
	assert.Equal(t, " branch.ab +0 -0", statusLines[3].(git.CommentStatusLine).Comment)
}
