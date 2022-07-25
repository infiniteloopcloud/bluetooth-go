package bluetooth

import (
	"strconv"
	"strings"
)

type Params struct {
	Log Log
}

// stringToByteArray converts MAC address string representation to little-endian byte array
func stringToByteArray(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}
