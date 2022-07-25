package bluetooth

import (
	"fmt"
	"strconv"
	"strings"
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

type SockaddrBth struct {
	family         uint16
	BtAddr         uint64
	ServiceClassId windows.GUID
	Port           uint32
}

func (sa *SockaddrBth) sockaddr() (unsafe.Pointer, int32, error) {
	// if sa.Port < 0 || sa.Port > 31 {
	// 	return nil, 0, windows.EINVAL
	// }
	sa.family = windows.AF_BTH
	p := (*[2]byte)(unsafe.Pointer(&sa.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	fmt.Println(" --- SockaddrBth: ", unsafe.Sizeof(SockaddrBth{}))
	fmt.Println(" --- family: ", unsafe.Sizeof(*&sa.family))
	fmt.Println(" --- BtAddr: ", unsafe.Sizeof(*&sa.BtAddr))
	fmt.Println(" --- ServiceClassId: ", unsafe.Sizeof(*&sa.ServiceClassId))
	fmt.Println(" --- Port: ", unsafe.Sizeof(*&sa.Port))
	fmt.Println(" --- SizeOf: ", unsafe.Sizeof(*sa))
	fmt.Println(" --- SizeOf int32: ", int32(unsafe.Sizeof(*sa)))
	upsa := unsafe.Pointer(sa)
	return upsa, int32(unsafe.Sizeof(*sa)), nil
}

func Bluetooth(addr string, params Params) (Communicator, error) {
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
	s := SockaddrBth{
		BtAddr: 0x54812D7FCDD2,
		//ServiceClassId: g,
		Port: 6,
	}
	if err := Connect(fd, s); err != nil {
		return nil, err
	}
	params.Log.Print("unix socket linked with an RFCOMM")

	return &bluetooth{
		log:    params.Log,
		Handle: fd,
		Addr:   addr,
	}, nil
}

func (b *bluetooth) Read(dataLen int) (int, []byte, error) {

	var data = make([]byte, dataLen)
	n, err := windows.Read(b.Handle, data)
	if err != nil {
		return 0, nil, err
	}
	b.log.Print(fmt.Sprintf(">>>>>>>>>>>> protoComm.Read: %v", data[:n]))
	return 12, data, nil
}

func (b *bluetooth) Write(d []byte) error {
	b.log.Print(">>>>>>>>>>>> protoComm.Write: %v\n", d)
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

// str2ba converts MAC address string representation to little-endian byte array
func str2ba(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}

// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------
/// Temporary
// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------

func Connect(fd windows.Handle, sa SockaddrBth) (err error) {
	// ptr, n, err := sa.sockaddr()
	// if err != nil {
	// 	return err
	// }

	// 20 00 d2 cd 7f 2d 81 54 00 00 7f 21 96 00 00 00 54 00 00 00 81 00 00 00 2d 00 06 00 00 00
	ptr := unsafe.Pointer(&[30]byte{0x20, 0x00, 0xd2, 0xcd, 0x7f, 0x2d, 0x81, 0x54, 0x00, 0x00, 0x7f, 0x21, 0x96, 0x00, 0x00, 0x00, 0x54,
		0x00, 0x00, 0x00, 0x81, 0x00, 0x00, 0x00, 0x2d, 0x00, 0x06, 0x00, 0x00, 0x00})
	return connectOg(fd, ptr, 30)
}

const socket_error = uintptr(^uint32(0))

var (
	modws2_32   = windows.NewLazySystemDLL("ws2_32.dll")
	procconnect = modws2_32.NewProc("connect")
	procsend    = modws2_32.NewProc("send")
)

func connectOg(s windows.Handle, name unsafe.Pointer, namelen int32) (err error) {
	r1, _, e1 := syscall.Syscall(procconnect.Addr(), 3, uintptr(s), uintptr(name), uintptr(namelen))
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
