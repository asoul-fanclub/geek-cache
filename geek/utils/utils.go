package utils

import (
	"strings"
)

func ValidPeerAddr(addr string) bool {
	t1 := strings.Split(addr, ":")
	if len(t1) != 2 {
		return false
	}
	// TODO: more selections
	t2 := strings.Split(t1[0], ".")
	if t1[0] != "localhost" && len(t2) != 4 {
		return false
	}
	return true
}
