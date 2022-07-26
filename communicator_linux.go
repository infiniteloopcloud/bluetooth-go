package bluetooth

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// NOTE: https://github.com/golang/go/issues/52325
var _ Communicator = &bluetooth{}

type bluetooth struct {
	log            Log
	FileDescriptor int
	SocketAddr     *unix.SockaddrRFCOMM
	Addr           string
}

func New(addr string, params Params) (Communicator, error) {
	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return nil, err
	}
	params.Log.Print("unix socket returned a file descriptor: ", fd)
	socketAddr := &unix.SockaddrRFCOMM{Addr: addressToByteArray(addr), Channel: 6}
	if err := unix.Connect(fd, socketAddr); err != nil {
		return nil, err
	}
	params.Log.Print("unix socket linked with an RFCOMM")

	return &bluetooth{
		log:            params.Log,
		FileDescriptor: fd,
		SocketAddr:     socketAddr,
		Addr:           addr,
	}, nil
}

func (b *bluetooth) Read(dataLen int) (int, []byte, error) {
	var data = make([]byte, dataLen)
	n, err := unix.Read(b.FileDescriptor, data)
	if err != nil {
		return 0, nil, err
	}
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Read: %v", data[:n]))
	return n, data, nil
}

func (b *bluetooth) Write(d []byte) error {
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Write: %v", d))
	_, err := unix.Write(b.FileDescriptor, d)
	if err != nil {
		return err
	}
	return nil
}

func (b bluetooth) Close() error {
	return unix.Close(b.FileDescriptor)
}
