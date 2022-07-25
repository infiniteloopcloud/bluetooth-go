package bluetooth

import (
	"strconv"
	"strings"
)

func Get(addr string, params Params) (Communicator, error) {
	return New(addr, params)
}

type Params struct {
	Log Log
}

// addressToByteArray converts MAC address string representation to little-endian byte array
func addressToByteArray(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}
