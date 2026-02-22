package fileutil

import (
	"regexp"
	"strings"
)

// ReplaceInFile replaces the first occurrence of a string or regex in a file.
// Returns true if the file was modified.
func ReplaceInFile(filePath string, search string, replace string) bool {
	content, ok := ReadFile(filePath)
	if !ok {
		return false
	}

	newContent := strings.Replace(content, search, replace, 1)
	if newContent == content {
		return false
	}

	return WriteFile(filePath, newContent) == nil
}

// ReplaceAllInFile replaces all occurrences of a string in a file.
func ReplaceAllInFile(filePath string, search string, replace string) bool {
	content, ok := ReadFile(filePath)
	if !ok {
		return false
	}

	newContent := strings.ReplaceAll(content, search, replace)
	if newContent == content {
		return false
	}

	return WriteFile(filePath, newContent) == nil
}

// ReplaceRegexInFile replaces the first match of a regex pattern in a file.
func ReplaceRegexInFile(filePath string, pattern string, replace string) bool {
	content, ok := ReadFile(filePath)
	if !ok {
		return false
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	newContent := re.ReplaceAllString(content, replace)
	if newContent == content {
		return false
	}

	return WriteFile(filePath, newContent) == nil
}




