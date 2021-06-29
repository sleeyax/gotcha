package utils

import (
	urlPkg "net/url"
	"strings"
)

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

// MergeUrl computes the actual request url by combining prefixUrl and url.
// If both prefixUrl and url are absolute, gotcha will assume url to be the root url.
func MergeUrl(prefixUrl string, url string) (*urlPkg.URL, error) {
	if prefixUrl == "" {
		return urlPkg.Parse(url)
	}

	pu, err := urlPkg.Parse(prefixUrl)
	if err != nil {
		return nil, err
	}

	if url == "" {
		return pu, err
	}

	u, err := urlPkg.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.IsAbs() {
		return u, nil
	}

	return pu.Parse(url)
}
