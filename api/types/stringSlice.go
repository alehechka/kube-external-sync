package types

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// StringSlice is a slice of strings
type StringSlice []string

// IsExcluded determines whether or not a given string is excluded in terms of a blacklist StringSlice.
// If the StringSlice is empty, then all provided strings are considered not excluded.
// If the provided string exists within the StringSlice, then it is considered excluded.
func (slice StringSlice) IsExcluded(str string) bool {
	if slice.IsEmpty() {
		return false
	}

	return slice.exists(str)
}

// IsIncluded determines whether or not a given string is included in terms of a whitelist StringSlice.
// If the StringSlice is empty, then all provided strings are considered included.
// If the StringSlice is not empty, then only strings that exist in the StringSlice will be considered included.
func (slice StringSlice) IsIncluded(str string) bool {
	return slice.exists(str)
}

func (slice StringSlice) exists(str string) bool {
	for _, obj := range slice {
		if obj == str {
			return true
		}
	}

	return false
}

// IsEmpty checks if length of slice is greater than zero
func (slice StringSlice) IsEmpty() bool {
	return len(slice) == 0
}

// IsRegexExcluded determines whether or not a given string is excluded in terms of a blacklist RegexSlice of regular expressions.
// If the RegexSlice is empty, then all provided strings are considered not excluded.
// If the provided string exists within the RegexSlice regular expressions, then it is considered excluded.
func (slice StringSlice) IsRegexExcluded(str string) bool {
	if slice.IsEmpty() {
		return false
	}

	return slice.regexExists(str)
}

// IsRegexIncluded determines whether or not a given string is included in terms of a whitelist RegexSlice of regular expressions.
// If the RegexSlice is empty, then all provided strings are considered included.
// If the RegexSlice is not empty, then only strings that exist in the RegexSlice regular expressions will be considered included.
func (slice StringSlice) IsRegexIncluded(str string) bool {
	return slice.regexExists(str)
}

func (slice StringSlice) regexExists(str string) bool {
	for _, pattern := range slice {
		if match, err := regexp.MatchString(pattern, str); match {
			return true
		} else if err != nil {
			log.WithFields(log.Fields{"pattern": pattern}).Errorf(err.Error())
		}
	}

	return false
}
