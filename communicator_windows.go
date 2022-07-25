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

func New(addr string, params Params) (Communicator, error) {
	var d syscall.WSAData
	e := syscall.WSAStartup(uint32(0x202), &d)
	if e != nil {
		return nil, e
	}

	fd, err := windows.Socket(windows.AF_BTH, windows.SOCK_STREAM, windows.BTHPROTO_RFCOMM)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	// 54:81:2D:7F:CD:D2
	s := SockaddrBth{
		BtAddr: 0x54812d7fcdd2,
		Port:   6,
	}
	if err := Connect(fd, s); err != nil {
		return nil, err
	}
	params.Log.Print("unix socket linked with an RFCOMM")

	return &bluetooth{
		//FileDescriptor: fd,
		//SocketAddr:     socketAddr,
		log:    params.Log,
		Handle: fd,
		Addr:   addr,
	}, nil
}

func (b *bluetooth) Read(dataLen int) (int, []byte, error) {
	var Length [500]byte
	UitnZero_1 := uint32(0)

	buf := windows.WSABuf{Len: uint32(500), Buf: &Length[0]}
	recv := uint32(0)
	err := windows.WSARecv(b.Handle, &buf, 1, &recv, &UitnZero_1, nil, nil)
	if err != nil {
		return 0, nil, err
	}
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Read: %v", Length[:recv]))
	var data = make([]byte, recv)
	for i := 0; i < int(recv); i++ {
		data[i] = Length[i]
	}

	return int(recv), data, nil
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

const socket_error = uintptr(^uint32(0))

var (
	modws2_32   = windows.NewLazySystemDLL("ws2_32.dll")
	procconnect = modws2_32.NewProc("connect")
	procsend    = modws2_32.NewProc("send")
)

func connectOg(s windows.Handle, name unsafe.Pointer, namelen int32) (err error) {
	r1, _, e1 := syscall.SyscallN(procconnect.Addr(), 3, uintptr(s), uintptr(name), uintptr(namelen))
	if r1 == socket_error {
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
