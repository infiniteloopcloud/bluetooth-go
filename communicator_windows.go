package bluetooth

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

var _ Communicator = &bluetooth{}

type bluetooth struct {
	log    Log
	Handle windows.Handle
	Addr   string
}

func Connect(params Params) (Communicator, error) {
	var d syscall.WSAData
	err := syscall.WSAStartup(uint32(0x202), &d)
	if err != nil {
		return nil, err
	}

	fd, err := windows.Socket(windows.AF_BTH, windows.SOCK_STREAM, windows.BTHPROTO_RFCOMM)
	if err != nil {
		return &bluetooth{
			log:    params.Log,
			Handle: fd,
			Addr:   params.Address,
		}, err
	}

	addressUint64, err := addressToUint64(params.Address)
	if err != nil {
		return &bluetooth{
			log:    params.Log,
			Handle: fd,
			Addr:   params.Address,
		}, err
	}
	s := &windows.SockaddrBth{
		BtAddr: addressUint64,
		Port:   6,
	}
	if err := windows.Connect(fd, s); err != nil {
		return &bluetooth{
			log:    params.Log,
			Handle: fd,
			Addr:   params.Address,
		}, err
	}
	params.Log.Print("unix socket linked with an RFCOMM")

	return &bluetooth{
		log:    params.Log,
		Handle: fd,
		Addr:   params.Address,
	}, nil
}

func (b *bluetooth) Read(dataLen int) (int, []byte, error) {
	var data = make([]byte, dataLen)
	flags := uint32(0)

	buf := windows.WSABuf{Len: uint32(dataLen), Buf: &data[0]}
	receiver := uint32(0)
	err := windows.WSARecv(b.Handle, &buf, 1, &receiver, &flags, nil, nil)
	if err != nil {
		return 0, nil, err
	}
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Read: %v", data[:receiver]))

	return int(receiver), data, nil
}

func (b *bluetooth) Write(d []byte) (int, error) {
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Write: %v\n", d))
	buf := &windows.WSABuf{
		Len: uint32(len(d)),
	}
	if len(d) > 0 {
		buf.Buf = &d[0]
	}
	var numOfBytes uint32
	err := windows.WSASend(b.Handle, buf, 1, &numOfBytes, 0, nil, nil)
	return int(numOfBytes), err
}

func (b bluetooth) Close() error {
	return windows.Close(b.Handle)
}
