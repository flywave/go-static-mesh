package utils

import "strings"

func HasPrefix(s, prefix string) (bool, string) {
	if strings.HasPrefix(s, prefix) {
		return true, s[len(prefix):]
	}
	return false, ""
}
