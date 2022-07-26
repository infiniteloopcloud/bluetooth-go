package bluetooth

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ Communicator = &bluetooth{}

type bluetooth struct {
	log    Log
	Handle windows.Handle
	Addr   string
}

func New(params Params) (Communicator, error) {
	var d syscall.WSAData
	err := syscall.WSAStartup(uint32(0x202), &d)
	if err != nil {
		return nil, err
	}

	fd, err := windows.Socket(windows.AF_BTH, windows.SOCK_STREAM, windows.BTHPROTO_RFCOMM)
	if err != nil {
		return nil, err
	}

	addressUint64, err := addressToUint64(params.Address)
	if err != nil {
		return nil, err
	}
	s := SockaddrBth{
		BtAddr: addressUint64,
		Port:   6,
	}
	if err := Connect(fd, s); err != nil {
		return nil, err
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

func (b *bluetooth) Write(d []byte) error {
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Write: %v\n", d))
	buf := &windows.WSABuf{
		Len: uint32(len(d)),
	}
	if len(d) > 0 {
		buf.Buf = &d[0]
	}
	var numOfBytes uint32
	err := windows.WSASend(b.Handle, buf, 1, &numOfBytes, 0, nil, nil)
	return err
}

func (b bluetooth) Close() error {
	return windows.Close(b.Handle)
}

// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------
/// Temporary
// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------

func Connect(fd windows.Handle, sa SockaddrBth) (err error) {
	ptr, n, err := sa.sockaddr()
	if err != nil {
		return err
	}

	return connectOg(fd, ptr, n)
}

func connectOg(s windows.Handle, name unsafe.Pointer, namelen int32) (err error) {
	procConnection := windows.NewLazySystemDLL("ws2_32.dll").NewProc("connect")
	r1, _, e1 := syscall.SyscallN(procConnection.Addr(), 3, uintptr(s), uintptr(name), uintptr(namelen))
	if r1 == uintptr(^uint32(0)) {
		err = errnoErr(e1)
	}
	return
}

const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	return e
}

type RawSockaddrBth struct {
	AddressFamily  [2]byte
	btAddr         [8]byte
	ServiceClassId [16]byte
	Port           [4]byte
}

type SockaddrBth struct {
	BtAddr         uint64
	ServiceClassId windows.GUID
	Port           uint32

	raw RawSockaddrBth
}

func (sa *SockaddrBth) sockaddr() (unsafe.Pointer, int32, error) {
	family := windows.AF_BTH
	sa.raw = RawSockaddrBth{
		AddressFamily:  *(*[2]byte)(unsafe.Pointer(&family)),
		btAddr:         *(*[8]byte)(unsafe.Pointer(&sa.BtAddr)),
		Port:           *(*[4]byte)(unsafe.Pointer(&sa.Port)),
		ServiceClassId: *(*[16]byte)(unsafe.Pointer(&sa.ServiceClassId)),
	}
	return unsafe.Pointer(&sa.raw), int32(unsafe.Sizeof(sa.raw)), nil
}
