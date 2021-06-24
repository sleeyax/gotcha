package utils

import "strings"

// StringArrayContains checks is given string contains any of the provided values.
func StringArrayContains(values []string, str string) bool {
	for _, value := range values {
		if strings.Contains(str, value) {
			return true
		}
	}
	return false
}

// IntArrayContains checks is given int exists in any of the provided values.
func IntArrayContains(values []int, i int) bool {
	for _, value := range values {
		if i == value {
			return true
		}
	}
	return false
}
