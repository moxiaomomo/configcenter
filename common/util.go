package common

import (
	"strings"
)

func Basename(s string, withExtension bool) string {
	slash := strings.LastIndex(s, "/") // -1 if "/" not found
	s = s[slash+1:]
	if withExtension {
		return s
	}
	if dot := strings.LastIndex(s, "."); dot >= 0 {
		s = s[:dot]
	}
	return s
}
