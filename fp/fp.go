package fp

import (
	"path/filepath"
	"regexp"
	"strings"
)

func CleanFilePath(path string) string {
	inputParts := SplitFilePath(path)
	temp := make([]string, 0, len(inputParts))
	inRange := func(i int) bool {
		return i < len(inputParts)
	}
	separator := func(i int) bool {
		return inRange(i) && inputParts[i][0] == filepath.Separator
	}
	is := func(i int, s string) bool {
		return inRange(i) && inputParts[i] == s
	}
	for i := 0; i < len(inputParts); i++ {
		s := inputParts[i]
		if s == "." && separator(i+1) {
			i++
			continue
		}
		if !separator(i) && !is(i, "..") && separator(i+1) && is(i+2, "..") {
			if separator(i + 3) {
				i += 3
			} else {
				i += 2
			}
			continue
		}
		if separator(i) {
			temp = append(temp, string(filepath.Separator))
		} else {
			temp = append(temp, s)
		}
	}
	inputParts = temp
	if len(inputParts) == 0 {
		return "."
	}
	// Remove / at the end
	if inputParts[len(inputParts)-1] == "/" && len(inputParts) > 1 {
		inputParts = inputParts[:len(inputParts)-1]
	}
	// Remove /. at the end
	if inputParts[len(inputParts)-1] == "." && len(inputParts) > 2 && inputParts[len(inputParts)-2] == "/" {
		inputParts = inputParts[:len(temp)-2]
	}
	return strings.Join(inputParts, "")
}

// SplitFilePath
// Split a dirty file path, preserving it
// "" --> []
// "aa//aaa" --> ["aa", "//", "aaa"]
// "/aaaa" --> ["/", "aaa"]
// "/////" --> ["////"]
func SplitFilePath(path string) []string {
	lastSeparatorIndex := -1
	var result []string
	for i := 0; i < len(path); i++ {
		if path[i] != filepath.Separator {
			continue
		}
		count := countSeparators(path, i)
		if lastSeparatorIndex != -1 {
			result = append(result, path[lastSeparatorIndex+1:i])
		} else if i != 0 {
			result = append(result, path[0:i])
		}
		result = append(result, strings.Repeat(string(filepath.Separator), count))
		i += count - 1
		lastSeparatorIndex = i
	}
	if lastSeparatorIndex < len(path)-1 {
		// This works with lastSeparatorIndex==-1
		result = append(result, path[lastSeparatorIndex+1:])
	}
	if result == nil {
		return []string{}
	}
	return result
}

func ValidFilePathForOrto(path string) bool {
	if len(path) == 0 {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9_.\-~@#$%^&=+{}\[\]:;,<>]+$`)
	parts := SplitFilePath(filepath.Clean(path))
	for _, part := range parts {
		if part[0] != filepath.Separator {
			matches := re.FindStringSubmatch(part)
			if matches == nil {
				return false
			}
		}
	}
	return true
}

func countSeparators(path string, index int) int {
	count := 1
	for ; index+count < len(path) && path[index+count] == filepath.Separator; count++ {
	}
	return count
}
