package git

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
	xy   string
	sub  string
	path string
	mH   string
	mI   string
	mW   string
	hH   string
	hI   string
}
type RenamedOrCopiedStatusLine struct {
	change   ChangedStatusLine
	origPath string
	score    string
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

func validateChangedStatusLine(changedStatusLine ChangedStatusLine, line string) {
	if changedStatusLine.sub != "N..." {
		log.Fatal("Error parsing " + line + " (submodules not supported yet)")
	}
	if !IsValidGitMode(changedStatusLine.mH) {
		log.Fatal("Invalid git mode: " + changedStatusLine.mH + " in line: " + line)
	}
	if !IsValidGitMode(changedStatusLine.mI) {
		log.Fatal("Invalid git mode: " + changedStatusLine.mI + " in line: " + line)
	}
	if !IsValidGitMode(changedStatusLine.mW) {
		log.Fatal("Invalid git mode: " + changedStatusLine.mW + " in line: " + line)
	}
	if !IsSupportedGitMode(changedStatusLine.mH) {
		log.Fatal("Unsupported git mode: " + changedStatusLine.mH + " in line: " + line)
	}
	if !IsSupportedGitMode(changedStatusLine.mI) {
		log.Fatal("Unsupported git mode: " + changedStatusLine.mI + " in line: " + line)
	}
	if !IsSupportedGitMode(changedStatusLine.mW) {
		log.Fatal("Unsupported git mode: " + changedStatusLine.mW + " in line: " + line)
	}
	if fp.ChecksumGetAlgo(changedStatusLine.hH) == fp.UNKNOWN {
		log.Fatal("Unknown checksum algorithm: " + changedStatusLine.hH + " in line: " + line)
	}
	if fp.ChecksumGetAlgo(changedStatusLine.hI) == fp.UNKNOWN {
		log.Fatal("Unknown checksum algorithm: " + changedStatusLine.hI + " in line: " + line)
	}
	fp.ValidFilePathForOrtoOrDie(changedStatusLine.path)
}

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

// Unmerged entries have the following format; the first character is a "u" to distinguish from ordinary changed entries.
//
// u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
//
// Field       Meaning
// --------------------------------------------------------
// <XY>        A 2 character field describing the conflict type
// 	 as described in the short format.
// <sub>       A 4 character field describing the submodule state
// as described above.
// <m1>        The octal file mode in stage 1.
// <m2>        The octal file mode in stage 2.
// <m3>        The octal file mode in stage 3.
// <mW>        The octal file mode in the worktree.
// <h1>        The object name in stage 1.
// <h2>        The object name in stage 2.
// <h3>        The object name in stage 3.
// <path>      The pathname.
//--------------------------------------------------------

// TODO:
//        Pathname Format Notes and -z
//           When the -z option is given, pathnames are printed as is and without any quoting and lines are terminated with a NUL (ASCII 0x00) byte.
//           Without the -z option, pathnames with "unusual" characters are quoted as explained for the configuration variable core.quotePath (see git-config(1)).

// TODO: use yield:
func (xxx *XXX) CheckForEmptyLine(line string) bool {
	xxx.lastLineIndex++
	if line != "" {
		return true
	}
	if xxx.emptyLineIndex != 0 {
		log.Fatal("More than one empty line in git output")
	}
	xxx.emptyLineIndex = xxx.lastLineIndex
	return false
}

func (xxx *XXX) CheckForEmptyLineEnd() {
	if xxx.emptyLineIndex != 0 && xxx.emptyLineIndex != xxx.lastLineIndex {
		log.Fatal("Empty line not at the end in git output")
	}
}

func JoinInputWhenNeeded(line *string, prevLine *string) bool {
	if strings.HasPrefix(*line, "2 ") {
		*prevLine = *line
		return false
	}
	if *prevLine != "" {
		*line = *prevLine + "\x00" + *line
		*prevLine = ""
	}
	return true
}

func JoinInputWhenNeededEnd(prevLine *string) {
	if *prevLine != "" {
		log.Fatal("Incomplete line in git output")
	}
}

type XXX struct {
	emptyLineIndex int
	lastLineIndex  int
}

func RunStatus() []StatusLine {
	cmd := exec.Command("git", "status", "--porcelain=v2", "--untracked-files=all", "--show-stash", "--branch", "--ignored", "-z")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	var result []StatusLine
	output := string(out)
	lines := strings.SplitSeq(output, "\x00")
	var prevLine string
	var xxx XXX
	for line := range lines {
		if xxx.CheckForEmptyLine(line) && JoinInputWhenNeeded(&line, &prevLine) {
			result = append(result, parseLine(line))
			println(line)
		}
	}
	xxx.CheckForEmptyLineEnd()
	JoinInputWhenNeededEnd(&prevLine)
	log.Printf("%#v\n", result)
	return result
}

// TODO: even though it's using \x00, it's not receiving NUL bytes.
func parseLine(line string) StatusLine {
	if strings.HasPrefix(line, "#") {
		return CommentStatusLine{comment: line[1:]}
	} else if strings.HasPrefix(line, "! ") {
		fp.ValidFilePathForOrtoOrDie(line[2:])
		return IgnoredStatusLine{Path: line[2:]}
	} else if strings.HasPrefix(line, "? ") {
		fp.ValidFilePathForOrtoOrDie(line[2:])
		return UntrackedStatusLine{path: line[2:]}
	} else if strings.HasPrefix(line, "u ") {
		// TODO: this is untested and incomplete:
		// u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
		re := regexp.MustCompile(`^u \s*(\S{2})\s+(\S{4})\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+([0-9a-fA-F]+)\s+([0-9a-fA-F]+)\s+([0-9a-fA-F]+)\s+([^\x00]+)$`)
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			log.Fatal("Error parsing " + line)
		}
		fp.ValidFilePathForOrtoOrDie(matches[10])
		return UnmergedStatusLine{path: matches[10]}
	} else if strings.HasPrefix(line, "1 ") {
		// 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
		//1 .M N... 100644 100644 100644 e424a2f681538c6794e104ae2118919b3a2b74ef e424a2f681538c6794e104ae2118919b3a2b74ef git/status.go
		re := regexp.MustCompile(`^1 (\S{2}) (\S{4}) (\d+) (\d+) (\d+) ([0-9a-fA-F]+) ([0-9a-fA-F]+) ([^\x00]+)$`)
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			log.Fatal("Error parsing " + line)
		}

		changedStatusLine := ChangedStatusLine{xy: matches[1],
			sub:  matches[2],
			mH:   matches[3],
			mI:   matches[4],
			mW:   matches[5],
			hH:   matches[6],
			hI:   matches[7],
			path: matches[8]}
		validateChangedStatusLine(changedStatusLine, line)
		//log.Printf("%#v\n", changedStatusLine)
		return changedStatusLine
	} else if strings.HasPrefix(line, "2 ") {
		// 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
		// 2 R. N... 100644 100644 100644 37ee65c344d8ab16aebbed88699b77f3a0f2ee7f 37ee65c344d8ab16aebbed88699b77f3a0f2ee7f R100 git/blob.go   git/gitfile.go
		re := regexp.MustCompile(`^2 (\S{2}) (\S{4}) (\d+) (\d+) (\d+) ([0-9a-fA-F]+) ([0-9a-fA-F]+) ([CR]\d+) ([^\x00]+)\x00([^\x00]+)$`)
		matches := re.FindStringSubmatch(line)
		changedStatusLine := ChangedStatusLine{xy: matches[1],
			sub:  matches[2],
			mH:   matches[3],
			mI:   matches[4],
			mW:   matches[5],
			hH:   matches[6],
			hI:   matches[7],
			path: matches[9]}
		validateChangedStatusLine(changedStatusLine, line)
		renamedOrCopiedStatusLine := RenamedOrCopiedStatusLine{score: matches[8], origPath: matches[10], change: changedStatusLine}
		if matches == nil {
			log.Fatal("Error parsing " + line)
		}
		fp.ValidFilePathForOrtoOrDie(renamedOrCopiedStatusLine.origPath)
		log.Printf("%#v\n", renamedOrCopiedStatusLine)
		return renamedOrCopiedStatusLine
		// u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
	} else {
		panic("Status line can't be parsed: " + line)
	}
}
