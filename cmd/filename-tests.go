package main

// TODO: https://stackoverflow.com/questions/6119956/how-to-determine-if-git-handles-a-file-as-binary-or-as-text

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func closeKill(file *os.File, path string) {
	_ = os.Remove(path) // Will immediately remove the file in macos/UNIX
	_ = file.Close()
	_ = os.Remove(path) // In Windows, deletion only works here.
}

func tryCreateUnlinked(dir, name string) error {
	p := filepath.Join(dir, name)

	// O_EXCL ensures no accidental overwrite if a file appears between checks.
	file, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		// collision, bail out
		if os.IsExist(err) {
			println("File '" + name + "' already exists! Test inconclusive, bailing out.")
			os.Exit(2)
		}
		return err
		// EINVAL/ENAMETOOLONG/EPERM
	}
	closeKill(file, p)
	return nil
}

func tryCreateTwo(dir, name1, name2 string) error {
	p1 := filepath.Join(dir, name1)
	p2 := filepath.Join(dir, name2)

	// O_EXCL ensures no accidental overwrite if a file appears between checks.
	file1, err1 := os.OpenFile(p1, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err1 != nil {
		println("Error trying to create '" + p1 + "' Test inconclusive, bailing out.")
		os.Exit(2)
	}
	file2, err2 := os.OpenFile(p2, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	closeKill(file1, p1)

	if err2 != nil {
		return err2
	}
	closeKill(file2, p2)
	return nil
}

type NameTestResult int

const (
	NameTestOk NameTestResult = iota
	NameTestErrNameTooLong
	NameTestErrIllegalNameSequence
	NameTestErrInvalidArgument
)

func testName(name string) NameTestResult {
	err := tryCreateUnlinked("./", name)
	if err == nil {
		return NameTestOk
	}
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		if errors.Is(pathError.Err, syscall.ENAMETOOLONG) {
			return NameTestErrNameTooLong
		} else if errors.Is(pathError.Err, syscall.EILSEQ) {
			return NameTestErrIllegalNameSequence
		} else if errors.Is(pathError.Err, syscall.EINVAL) {
			return NameTestErrInvalidArgument
		} else if pathError.Err.Error() == "The parameter is incorrect." ||
			pathError.Err.Error() == "The system cannot find the file specified." ||
			pathError.Err.Error() == "The filename, directory name, or volume label syntax is incorrect." { // TODO: this is a windows thing
			return NameTestErrIllegalNameSequence // 0x57?
		} else {
			println(pathError.Err.Error())
			fmt.Printf("Unknown error type: %#v\n", pathError)
			panic("Unknown error returned")
		}
	}
	println(pathError.Err.Error())
	fmt.Printf("Unknown error type: %T\n", err)
	panic("Unknown error returned")
}

func findLongestFilename(baseString string) int {
	limit := 1024
	for i := 1; i <= limit; i++ {
		test := strings.Repeat(baseString, i)
		result := testName(test)
		switch result {
		case NameTestOk:
		case NameTestErrNameTooLong, NameTestErrIllegalNameSequence:
			return i - 1
		default:
			panic("Unknown error returned: " + fmt.Sprintf("%d", result))
		}
	}
	return limit
}

func testLength() {
	println("Testing length up to 1024...")
	singleByte := "a"
	doubleByte := "\u00A9"   // Copyright symbol
	tripleByte := "\u2600"   // Sun
	quadByte := "\U0001F355" // Pizza
	//goland:noinspection GoBoolExpressions
	if len(singleByte) != 1 || len(doubleByte) != 2 || len(tripleByte) != 3 || len(quadByte) != 4 {
		panic("invalid byte lengths")
	}
	longest1 := findLongestFilename(singleByte)
	println(" Longest filename for a single byte string:", longest1, " (", len(singleByte)*longest1, "bytes)")
	longest2 := findLongestFilename(doubleByte)
	println(" Longest filename for double byte UTF-8:", longest2, " (", len(doubleByte)*longest2, "bytes)")
	longest3 := findLongestFilename(tripleByte)
	println(" Longest filename for triple byte UTF-8 (1 x UTF-16):", longest3, "(", len(tripleByte)*longest3, "bytes)")
	longest4 := findLongestFilename(quadByte)
	println(" Longest filename for quad byte UTF-8 (2 x UTF-16):", longest4, "(", len(quadByte)*longest4, "bytes)")
}

func testEncoding() {
	badUtf8 := string([]byte{0xC0, 0xAF})
	badUtf8Result := testName(badUtf8)
	switch badUtf8Result {
	case NameTestOk:
		println(" Invalid UTF-8 sequence was allowed")
	case NameTestErrIllegalNameSequence:
		println(" Invalid UTF-8 sequence was not allowed")
	default:
		panic("Unknown error returned: " + fmt.Sprintf("%d", badUtf8Result))
	}
}

func main() {
	println("Main separator character reported by go: '"+string(filepath.Separator)+"' #", filepath.Separator)

	for i := 0; i < 256; i++ {
		if os.IsPathSeparator(uint8(i)) {
			println("Go's os.IsPathSeparator() returns true for #", i)
		}
	}

	testCaseSensitivity()
	testStrings()
	testCharacters()
	testLength()
	testEncoding()
}

func testCaseSensitivity() {
	println("Testing case sensitivity")
	collection := [][2]string{{"abcdefg", "ABCDEFG"}, {"áéñ", "ÁÉÑ"}, {"á", "a"}}
	for _, s := range collection {
		a := s[0]
		b := s[1]
		err := tryCreateTwo("./", a, b)
		if err == nil {
			println("  Files are different: '" + a + "' and '" + b + "'")
		} else {
			if os.IsExist(err) {
				println("  Files are the same: '" + a + "' and '" + b + "'")
			} else {
				panic(err)
			}
		}
	}
}
func testBytesInLoop(f func(i int) string) {
	for i := 0; i < 256; i++ {
		if os.IsPathSeparator(uint8(i)) {
			continue
		}
		name := f(i)
		result := testName(name)
		if result != NameTestOk {
			character := ""
			if i >= 32 && i != 127 {
				character = " '" + string(byte(i)) + "'"
			}
			println("  Character rejected: #" + strconv.Itoa(i) + character)
		}
	}
}

func testCharacters() {
	println("  Testing characters #0 to #255 in the middle of a file name, excluding separator")
	testBytesInLoop(func(i int) string {
		return "file-" + string(byte(i)) + "a.txt"
	})
	println("  Testing characters #0 to #255 at the start of a file name, excluding separator")
	testBytesInLoop(func(i int) string {
		return string(byte(i)) + "a"
	})
	println("  Testing characters #0 to #255 at the end of a file name, excluding separator")
	testBytesInLoop(func(i int) string {
		return "a" + string(byte(i))
	})
	println("Testing whitespace only string")
	whitespace := string([]byte{' ', '\t', '\n', '\r'})
	result := testName(whitespace)
	if result != NameTestOk {
		println("  Whitespace only filename rejected")
	} else {
		println("  Whitespace only filename worked")
	}
}

func testStrings() {
	println("Testing commonly rejected strings (CON, PRN, etc)")
	wordList := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8",
		"COM9", "COM¹", "COM²", "COM³", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9", "LPT¹",
		"LPT²", "LPT³", "CLOCK$", "CONFIG$"}
	var ok []string
	var rejected []string
	for _, suffix := range []string{"", ".txt"} {
		for _, word := range wordList {
			word = word + suffix
			result := testName(word)
			if result != NameTestOk {
				rejected = append(rejected, word)
			} else {
				ok = append(ok, word)
			}
		}
	}
	if len(rejected) == 0 {
		println("  Rejected: None")
	} else {
		println("  Rejected: " + strings.Join(rejected, ", "))
	}
	if len(ok) == 0 {
		println("  OK: None")
	} else {
		println("  OK: " + strings.Join(ok, ", "))
	}
}

// TODO: Note that, in macOS, the ":" character is displayed as a "/"!

// Results in macOS:
//	Main separator character reported by go: '/' # 47
//	Go's os.IsPathSeparator() returns true for # 47
//	Testing case sensitivity
//	Files are the same: 'abcdefg' and 'ABCDEFG'
//	Files are the same: 'áéñ' and 'ÁÉÑ'
//	Files are different: 'á' and 'a'
//	Testing commonly rejected strings (CON, PRN, etc)
//	Rejected: None
//	OK: CON, PRN, AUX, NUL, COM1, COM2, COM3, COM4, COM5, COM6, COM7, COM8, COM9, COM¹, COM², COM³, LPT1, LPT2, LPT3, LPT4, LPT5, LPT6, LPT7, LPT8, LPT9, LPT¹, LPT², LPT³, CLOCK$, CONFIG$, CON.txt, PRN.txt, AUX.txt, NUL.txt, COM1.txt, COM2.txt, COM3.txt, COM4.txt, COM5.txt, COM6.txt, COM7.txt, COM8.txt, COM9.txt, COM¹.txt, COM².txt, COM³.txt, LPT1.txt, LPT2.txt, LPT3.txt, LPT4.txt, LPT5.txt, LPT6.txt, LPT7.txt, LPT8.txt, LPT9.txt, LPT¹.txt, LPT².txt, LPT³.txt, CLOCK$.txt, CONFIG$.txt
//	Testing characters #0 to #255 in the middle of a file name, excluding separator
//	Character rejected: #0
//	Testing characters #0 to #255 at the start of a file name, excluding separator
//	Character rejected: #0
//	Testing characters #0 to #255 at the end of a file name, excluding separator
//	Character rejected: #0
//	Testing whitespace only string
//	Whitespace only filename worked
//	Testing length up to 1024...
//	Longest filename for a single byte string: 255  ( 255 bytes)
//	Longest filename for double byte UTF-8: 255  ( 510 bytes)
//	Longest filename for triple byte UTF-8 (1 x UTF-16): 255 ( 765 bytes)
//	Longest filename for quad byte UTF-8 (2 x UTF-16): 127 ( 508 bytes)
//	Invalid UTF-8 sequence was not allowed
