package orto

import (
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/anknetau/orto/fp"
)

// TODO:
// https://git-scm.com/docs/git-status/2.31.0
// Pathname Format Notes and -z
// When the -z option is given, pathnames are printed as is and without any quoting and lines are terminated with a NUL (ASCII 0x00) byte.
// Without the -z option, pathnames with "unusual" characters are quoted as explained for the configuration variable core.quotePath (see git-config[1]).

type StatusLineType int

const (
	IgnoredStatusLineType StatusLineType = iota
	UntrackedStatusLineType
	CommentStatusLineType
	ChangedStatusLineType
	RenamedOrCopiedStatusLineType
	UnmergedStatusLineType
)

type StatusLine interface {
	Type() StatusLineType
}

type IgnoredStatusLine struct {
	Path string
}

type UntrackedStatusLine struct {
	path string
}

type CommentStatusLine struct {
	comment string
}
type ChangedStatusLine struct {
	path string
}
type RenamedOrCopiedStatusLine struct {
	path string
}
type UnmergedStatusLine struct {
	path string
}

func (IgnoredStatusLine) Type() StatusLineType         { return IgnoredStatusLineType }
func (UntrackedStatusLine) Type() StatusLineType       { return UntrackedStatusLineType }
func (CommentStatusLine) Type() StatusLineType         { return CommentStatusLineType }
func (ChangedStatusLine) Type() StatusLineType         { return ChangedStatusLineType }
func (RenamedOrCopiedStatusLine) Type() StatusLineType { return RenamedOrCopiedStatusLineType }
func (UnmergedStatusLine) Type() StatusLineType        { return UnmergedStatusLineType }

//goland:noinspection GrazieInspection
func GitRunStatus() []StatusLine {
	cmd := exec.Command("git", "status", "--porcelain=v2", "--untracked-files=all", "--show-stash", "--branch", "--ignored")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Paths returned here need to all be checked with fp.ValidFilePathForOrtoOrDie(path)
	var result []StatusLine
	output := string(out)
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")
	for line := range lines {
		if comment, found := strings.CutPrefix(line, "#"); found {
			// TODO: do something with this:
			result = append(result, &CommentStatusLine{comment: comment})
			continue
		}
		if path, found := strings.CutPrefix(line, "! "); found {
			fp.ValidFilePathForOrtoOrDie(path)
			result = append(result, &IgnoredStatusLine{Path: path})
		} else if path, found := strings.CutPrefix(line, "? "); found {
			fp.ValidFilePathForOrtoOrDie(path)
			result = append(result, &UntrackedStatusLine{path: path})
		} else if leftover, found := strings.CutPrefix(line, "1 "); found {
			// 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
			// Field       Meaning
			//--------------------------------------------------------
			//<XY>        A 2 character field containing the staged and
			//	    unstaged XY values described in the short format,
			//	    with unchanged indicated by a "." rather than
			//	    a space.
			//<sub>       A 4 character field describing the submodule state.
			//	    "N..." when the entry is not a submodule.
			//	    "S<c><m><u>" when the entry is a submodule.
			//	    <c> is "C" if the commit changed; otherwise ".".
			//	    <m> is "M" if it has tracked changes; otherwise ".".
			//	    <u> is "U" if there are untracked changes; otherwise ".".
			//<mH>        The octal file mode in HEAD.
			//<mI>        The octal file mode in the index.
			//<mW>        The octal file mode in the worktree.
			//<hH>        The object name in HEAD.
			//<hI>        The object name in the index.
			//<X><score>  The rename or CopyFile score (denoting the percentage
			//	    of similarity between the source and target of the
			//	    move or CopyFile). For example "R100" or "C75".
			//<path>      The pathname.  In a renamed/copied entry, this
			//	    is the target path.
			//<sep>       When the `-z` option is used, the 2 pathnames are separated
			//	    with a NUL (ASCII 0x00) byte; otherwise, a tab (ASCII 0x09)
			//	    byte separates them.
			//<origPath>  The pathname in the commit at HEAD or in the index.
			//	    This is only present in a renamed/copied entry, and
			//	    tells where the renamed/copied contents came from.
			//--------------------------------------------------------

			// X          Y     Meaning
			// -------------------------------------------------
			//	        [AMD]   not updated
			// M        [ MD]   updated in index
			// A        [ MD]   added to index
			// D                deleted from index
			// R        [ MD]   renamed in index
			// C        [ MD]   copied in index
			// [MARC]           index and work tree matches
			// [ MARC]     M    work tree changed since index
			// [ MARC]     D    deleted in work tree
			// [ D]        R    renamed in work tree
			// [ D]        C    copied in work tree
			// -------------------------------------------------
			// D           D    unmerged, both deleted
			// A           U    unmerged, added by us
			// U           D    unmerged, deleted by them
			// U           A    unmerged, added by them
			// D           U    unmerged, deleted by us
			// A           A    unmerged, both added
			// U           U    unmerged, both modified
			// -------------------------------------------------
			// ?           ?    untracked
			// !           !    ignored
			// -------------------------------------------------

			re := regexp.MustCompile(`^\s*(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)
			matches := re.FindStringSubmatch(leftover)
			if matches == nil {
				// TODO: what now?
			}

			//println("xy", matches[1])
			//println("sub", matches[2])
			//println("mH", matches[3])
			//println("mI", matches[4])
			//println("mW", matches[5])
			//println("hH", matches[6])
			//println("hI", matches[7])
			//println("path", matches[8])
			// TODO
			result = append(result, &ChangedStatusLine{path: matches[8]})
			//println(">>>>", line)
		} else if path, found := strings.CutPrefix(line, "2 "); found {
			fp.ValidFilePathForOrtoOrDie(path)
			// 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
			result = append(result, &RenamedOrCopiedStatusLine{path: path})
			// u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
		} else if path, found := strings.CutPrefix(line, "u "); found {
			fp.ValidFilePathForOrtoOrDie(path)
			result = append(result, &UnmergedStatusLine{path: path})
		} else {
			panic("Status line can't be parsed: " + line)
		}
	}
	return result
}
