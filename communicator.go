package bluetooth

import (
	"fmt"
	"strconv"
	"strings"
)

type Params struct {
	Address           string
	CharacteristicIDs []string
	Log               Log
	Verbose           bool
}

var _ Communicator = Printer{}

// Printer is a mock implementation of Communicator
type Printer struct {
	log Log
}

func (p Printer) Read(dataLen int) (int, []byte, error) {
	return dataLen, nil, nil
}

func (p Printer) Write(d []byte) (int, error) {
	p.log.Print(fmt.Sprintf("PRINTER %d >>> %v", len(d), d))
	return len(d), nil
}

func (p Printer) Close() error {
	return nil
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

// addressToUint64 converts MAC address string to uint64
func addressToUint64(address string) (uint64, error) {
	addressParts := strings.Split(address, ":")
	addressPartsLength := len(addressParts)
	var result uint64
	for i, tmp := range addressParts {
		u, err := strconv.ParseUint(tmp, 16, 8)
		if err != nil {
			return 0, err
		}
		push := 8 * (addressPartsLength - 1 - i)
		result += u << push
	}
	return result, nil
}
