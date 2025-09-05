package git

import (
	"fmt"
	"iter"
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

//go:generate go run golang.org/x/tools/cmd/stringer -type=StatusLineKind
type StatusLineKind int

type Status string

func NewStatus(s string) Status {
	if len(s) != 2 {
		log.Fatal("Invalid status string: " + s)
	}
	return Status(s)
}

const (
	StatusLineKindIgnored StatusLineKind = iota
	StatusLineKindUntracked
	StatusLineKindComment
	StatusLineKindChanged
	StatusLineKindRenamedOrCopied
	StatusLineKindUnmerged
)

type StatusLine interface {
	Kind() StatusLineKind
}

type IgnoredStatusLine struct {
	Path string
}

type UntrackedStatusLine struct {
	Path string
}

type CommentStatusLine struct {
	Comment string
}
type ChangedStatusLine struct {
	Path          string
	Status        Status
	Sub           string
	ModeHead      Mode
	ModeIndex     Mode
	ModeWorktree  Mode
	ChecksumHead  fp.Checksum
	ChecksumIndex fp.Checksum
}
type RenamedOrCopiedStatusLine struct {
	OrigPath string
	Score    string
	Change   ChangedStatusLine
}
type UnmergedStatusLine struct {
	Path           string
	Status         Status
	Sub            string
	ModeStage1     Mode
	ModeStage2     Mode
	ModeStage3     Mode
	ModeWorktree   Mode
	ChecksumStage1 fp.Checksum
	ChecksumStage2 fp.Checksum
	ChecksumStage3 fp.Checksum
}

func PrintStatusLine(line *StatusLine) {
	print((*line).Kind().String() + ": ")

	switch v := (*line).(type) {
	case IgnoredStatusLine:
		println("Path:" + v.Path)
	case UntrackedStatusLine:
		println("Path:" + v.Path)
	case CommentStatusLine:
		println("Comment:" + v.Comment)
	case ChangedStatusLine:
		fmt.Printf("%+v\n", v)
	case RenamedOrCopiedStatusLine:
		fmt.Printf("%+v\n", v)
	case UnmergedStatusLine:
		fmt.Printf("%+v\n", v)
	default:
		panic(fmt.Sprintf("Unhandled StatusLineKind: %#v", *line))
	}
}

func (IgnoredStatusLine) Kind() StatusLineKind         { return StatusLineKindIgnored }
func (UntrackedStatusLine) Kind() StatusLineKind       { return StatusLineKindUntracked }
func (CommentStatusLine) Kind() StatusLineKind         { return StatusLineKindComment }
func (ChangedStatusLine) Kind() StatusLineKind         { return StatusLineKindChanged }
func (RenamedOrCopiedStatusLine) Kind() StatusLineKind { return StatusLineKindRenamedOrCopied }
func (UnmergedStatusLine) Kind() StatusLineKind        { return StatusLineKindUnmerged }

func validateChangedStatusLine(changedStatusLine ChangedStatusLine, line string) {
	if changedStatusLine.Sub != "N..." {
		log.Fatal("Error parsing " + line + " (submodules not supported yet)")
	}
	fp.ValidFilePathForOrtoOrDie(changedStatusLine.Path)
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

// Pathname Format Notes and -z
//   When the -z option is given, pathnames are printed as is and without any quoting and lines are terminated with a NUL (ASCII 0x00) byte.
//   Without the -z option, pathnames with "unusual" characters are quoted as explained for the configuration variable core.quotePath (see git-config(1)).

func CheckForEmptyLineIter(input iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		emptyLineIndex := 0
		lastLineIndex := 0
		for line := range input {
			lastLineIndex++
			if line != "" {
				if !yield(line) {
					return
				}
				continue
			}
			if emptyLineIndex != 0 {
				log.Fatal("More than one empty line in git output")
			}
			emptyLineIndex = lastLineIndex
		}
		if emptyLineIndex != 0 && emptyLineIndex != lastLineIndex {
			log.Fatal("Empty line not at the end in git output")
		}
	}
}

func JoinInputWhenNeededIter(input iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		prevLine := ""
		for line := range input {
			if strings.HasPrefix(line, "2 ") {
				prevLine = line
				continue
			}
			if prevLine != "" {
				line = prevLine + "\x00" + line
				prevLine = ""
			}
			if !yield(line) {
				return
			}
		}
		if prevLine != "" {
			log.Fatal("Incomplete line in git output")
		}
	}
}

func RunStatus(config fp.EnvConfig) []StatusLine {
	cmd := exec.Command(config.GitCommand, "status", "--porcelain=v2", "--untracked-files=all", "--show-stash", "--branch", "--ignored", "-z")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: stream this stuff
	output := string(out)
	return ParseLines(output)
}

func ParseLines(output string) []StatusLine {
	var result []StatusLine
	lines := strings.SplitSeq(output, "\x00")
	for line := range CheckForEmptyLineIter(JoinInputWhenNeededIter(lines)) {
		result = append(result, ParseLine(line))
	}
	//log.Printf("%#v\n", result)
	return result
}

//goland:noinspection SpellCheckingInspection
const (
	xy    = `([MARCDU?!.]{2})`
	sub   = `(N[.]{3}|S[C.][M.][U.])`
	mode  = `([0-7]+)`
	hash  = `([0-9a-fA-F]+)`
	score = `([CR]\d+)`
	path  = `([^\x00]+)`
)

var (
	re1 = regexp.MustCompile("^1 " + xy + " " + sub + " " + mode + " " + mode + " " + mode + " " + hash + " " + hash + " " + path + "$")
	re2 = regexp.MustCompile("^2 " + xy + " " + sub + " " + mode + " " + mode + " " + mode + " " + hash + " " + hash + " " + score + " " + path + "\x00" + path + "$")
	reU = regexp.MustCompile("^u " + xy + " " + sub + " " + mode + " " + mode + " " + mode + " " + mode + " " + hash + " " + hash + " " + hash + " " + path + "$")
)

func ParseLine(line string) StatusLine {

	if strings.HasPrefix(line, "#") {
		return CommentStatusLine{Comment: line[1:]}
	} else if strings.HasPrefix(line, "! ") {
		fp.ValidFilePathForOrtoOrDie(line[2:])
		return IgnoredStatusLine{Path: line[2:]}
	} else if strings.HasPrefix(line, "? ") {
		fp.ValidFilePathForOrtoOrDie(line[2:])
		return UntrackedStatusLine{Path: line[2:]}
	} else if strings.HasPrefix(line, "u ") {
		// TODO: this is untested:
		// u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
		matches := reU.FindStringSubmatch(line)
		if matches == nil {
			log.Fatal("Error parsing " + line)
		}
		unmergedStatusLine := UnmergedStatusLine{
			Status:         NewStatus(matches[1]),
			Sub:            matches[2],
			ModeStage1:     NewMode(matches[3]),
			ModeStage2:     NewMode(matches[4]),
			ModeStage3:     NewMode(matches[5]),
			ModeWorktree:   NewMode(matches[6]),
			ChecksumStage1: fp.NewChecksum(matches[7]),
			ChecksumStage2: fp.NewChecksum(matches[8]),
			ChecksumStage3: fp.NewChecksum(matches[9]),
			Path:           matches[10],
		}
		fp.ValidFilePathForOrtoOrDie(unmergedStatusLine.Path)
		return unmergedStatusLine
	} else if strings.HasPrefix(line, "1 ") {
		// 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
		//1 .M N... 100644 100644 100644 e424a2f681538c6794e104ae2118919b3a2b74ef e424a2f681538c6794e104ae2118919b3a2b74ef git/status.go
		// 1 .D N... 100644 100644 000000 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 45b983be36b73c0788dc9cbcb76cbb80fc7bb057 deleteme
		matches := re1.FindStringSubmatch(line)
		if matches == nil {
			log.Fatal("Error parsing " + line)
		}

		changedStatusLine := ChangedStatusLine{
			Status:        NewStatus(matches[1]),
			Sub:           matches[2],
			ModeHead:      NewMode(matches[3]),
			ModeIndex:     NewMode(matches[4]),
			ModeWorktree:  NewMode(matches[5]),
			ChecksumHead:  fp.NewChecksum(matches[6]),
			ChecksumIndex: fp.NewChecksum(matches[7]),
			Path:          matches[8]}
		validateChangedStatusLine(changedStatusLine, line)
		//log.Printf("%#v\n", changedStatusLine)
		return changedStatusLine
	} else if strings.HasPrefix(line, "2 ") {
		// 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
		// 2 R. N... 100644 100644 100644 37ee65c344d8ab16aebbed88699b77f3a0f2ee7f 37ee65c344d8ab16aebbed88699b77f3a0f2ee7f R100 git/blob.go   git/gitfile.go
		matches := re2.FindStringSubmatch(line)
		changedStatusLine := ChangedStatusLine{
			Status:        NewStatus(matches[1]),
			Sub:           matches[2],
			ModeHead:      NewMode(matches[3]),
			ModeIndex:     NewMode(matches[4]),
			ModeWorktree:  NewMode(matches[5]),
			ChecksumHead:  fp.NewChecksum(matches[6]),
			ChecksumIndex: fp.NewChecksum(matches[7]),
			Path:          matches[9]}
		validateChangedStatusLine(changedStatusLine, line)
		renamedOrCopiedStatusLine := RenamedOrCopiedStatusLine{Score: matches[8], OrigPath: matches[10], Change: changedStatusLine}
		if matches == nil {
			log.Fatal("Error parsing " + line)
		}
		fp.ValidFilePathForOrtoOrDie(renamedOrCopiedStatusLine.OrigPath)
		log.Printf("%#v\n", renamedOrCopiedStatusLine)
		return renamedOrCopiedStatusLine
		// u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
	} else {
		panic("Status line can't be parsed: " + line)
	}
}
