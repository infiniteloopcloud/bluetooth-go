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
