package bluetooth

import "io"

type Log interface {
	Print(a ...interface{})
}

type Communicator interface {
	Read(dataLen int) (int, []byte, error)
	Write(d []byte) error
	io.Closer
}

type Device struct {
	Name       string `json:"name"`
	MACAddress string `json:"mac_address"`
}

type Scanner interface {
	Scan() []Device
}
