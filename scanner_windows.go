package bluetooth

import (
	"fmt"
	"sort"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

func NewScanner() Scanner {
	return scanner{}
}

var _ Scanner = scanner{}

type scanner struct{}

func (scanner) Scan() ([]Device, error) {
	var devices []Device
	var flags uint32 = LUP_CONTAINERS
	flags |= LUP_RETURN_NAME
	flags |= LUP_RETURN_ADDR

	var querySet WSAQUERYSET
	querySet.NameSpace = 16
	querySet.Size = uint32(unsafe.Sizeof(WSAQUERYSET{}))

	var handle windows.Handle
	err := WSALookupServiceBegin(&querySet, flags, &handle)
	if err != nil {
		return devices, err
	}

	var size = int32(unsafe.Sizeof(WSAQUERYSET{}))
	for i := 0; i < 5; i++ {
		var q WSAQUERYSET
		q, err = WSALookupServiceNext(handle, flags, &size)
		if err != nil {
			if strings.Contains(err.Error(), "No more results") {
				fmt.Printf("WSALookupServiceNext: %s\n", err.Error()) // TODO
				break
			}
			fmt.Printf("WSALookupServiceNext: %s\n", err.Error()) // TODO
		}

		devices = append(devices, recvDevice(&q))
	}

	err = WSALookupServiceEnd(handle)
	if err != nil {
		return devices, fmt.Errorf("WSALookupServiceEnd: %s", err.Error())
	}

	return devices, nil
}

func recvDevice(querySet *WSAQUERYSET) Device {
	if querySet == nil {
		return Device{}
	}
	var device Device
	if querySet.ServiceInstanceName != nil {
		var addr string
		for _, e := range querySet.SaBuffer.RemoteAddr.Sockaddr.Data {
			if e != 0 {
				addr += fmt.Sprintf("%x", e)
			}
		}
		device.MACAddress = parseRemoteSockaddr(addr)
		device.Name = querySet.ServiceInstanceNameToString()
	}
	return device
}

func parseRemoteSockaddr(addr string) string {
	var spaced = ""
	for i, elem := range []byte(addr) {
		if i%2 == 1 && i < 11 {
			spaced += string(elem) + " "
		} else {
			spaced += string(elem)
		}
	}
	addrSlice := strings.Split(spaced, " ")
	sort.SliceStable(addrSlice, func(i, j int) bool {
		return i > j
	})
	newAddr := strings.Join(addrSlice, ":")
	newAddr = strings.ToUpper(newAddr)
	return newAddr
}

// -------------------------------------------------
// TEMPORARY
// -------------------------------------------------

const (
	LUP_DEEP                = 0x0001
	LUP_CONTAINERS          = 0x0002
	LUP_NOCONTAINERS        = 0x0004
	LUP_NEAREST             = 0x0008
	LUP_RETURN_NAME         = 0x0010
	LUP_RETURN_TYPE         = 0x0020
	LUP_RETURN_VERSION      = 0x0040
	LUP_RETURN_COMMENT      = 0x0080
	LUP_RETURN_ADDR         = 0x0100
	LUP_RETURN_BLOB         = 0x0200
	LUP_RETURN_ALIASES      = 0x0400
	LUP_RETURN_QUERY_STRING = 0x0800
	LUP_RETURN_ALL          = 0x0FF0
	LUP_RES_SERVICE         = 0x8000

	LUP_FLUSHCACHE    = 0x1000
	LUP_FLUSHPREVIOUS = 0x2000

	LUP_NON_AUTHORITATIVE      = 0x4000
	LUP_SECURE                 = 0x8000
	LUP_RETURN_PREFERRED_NAMES = 0x10000
	LUP_DNS_ONLY               = 0x20000

	LUP_ADDRCONFIG           = 0x100000
	LUP_DUAL_ADDR            = 0x200000
	LUP_FILESERVER           = 0x400000
	LUP_DISABLE_IDN_ENCODING = 0x00800000
	LUP_API_ANSI             = 0x01000000

	LUP_RESOLUTION_HANDLE = 0x80000000
)

const socket_error = uintptr(^uint32(0))

const errnoERROR_IO_PENDING = 997

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

var (
	modws2_32                 = windows.NewLazySystemDLL("ws2_32.dll")
	procWSALookupServiceBegin = modws2_32.NewProc("WSALookupServiceBeginW")
	procWSALookupServiceNext  = modws2_32.NewProc("WSALookupServiceNextW")
	procWSALookupServiceEnd   = modws2_32.NewProc("WSALookupServiceEnd")
)

// https://docs.microsoft.com/en-us/windows/win32/api/winsock2/ns-winsock2-wsaquerysetw
type WSAQUERYSET struct {
	Size                uint32
	ServiceInstanceName *uint16
	ServiceClassId      *windows.GUID
	Version             *WSAVersion
	Comment             *uint16
	NameSpace           uint32
	NSProviderId        *windows.GUID
	Context             *uint16
	NumberOfProtocols   uint32
	AfpProtocols        *AFProtocols
	QueryString         *uint16
	NumberOfCsAddrs     uint32
	SaBuffer            *AddrInfo
	OutputFlags         uint32
	Blob                *BLOB
}

func (w WSAQUERYSET) ServiceInstanceNameToString() string {
	return RawPointerToString(w.ServiceInstanceName)
}

func (w WSAQUERYSET) CommentToString() string {
	return RawPointerToString(w.Comment)
}

func (w WSAQUERYSET) ContextToString() string {
	return RawPointerToString(w.Context)
}

func (w WSAQUERYSET) QueryStringToString() string {
	return RawPointerToString(w.QueryString)
}

// https://docs.microsoft.com/en-us/windows/win32/api/winsock2/ns-winsock2-wsaversion
type WSAVersion struct {
	Version                  uint32 // DWORD
	EnumerationOfComparision int32  // WSAEcomparator enum
}

// https://docs.microsoft.com/en-us/windows/win32/api/winsock2/ns-winsock2-afprotocols
type AFProtocols struct {
	AddressFamily int32
	Protocol      int32
}

// https://docs.microsoft.com/en-us/windows/win32/winsock/sockaddr-2
type Sockaddr struct {
	Family uint16
	Data   [14]byte
}

// https://docs.microsoft.com/en-us/windows/win32/api/Ws2def/ns-ws2def-socket_address
type SocketAddress struct {
	Sockaddr       *Sockaddr
	SockaddrLength int
}

// https://docs.microsoft.com/en-us/windows/win32/api/ws2def/ns-ws2def-csaddr_info
type AddrInfo struct {
	LocalAddr  SocketAddress
	RemoteAddr SocketAddress
	SocketType int32
	Protocol   int32
}

// https://docs.microsoft.com/en-us/windows/win32/api/winsock2/ns-winsock2-blob
type BLOB struct {
	Size     uint32
	BlobData *byte // TODO how to represent a block of data in Go?
}

func RawPointerToString(w *uint16) string {
	if w != nil {
		us := make([]uint16, 0, 256)
		for p := uintptr(unsafe.Pointer(w)); ; p += 2 {
			u := *(*uint16)(unsafe.Pointer(p))
			if u == 0 {
				return string(utf16.Decode(us))
			}
			us = append(us, u)
		}
	}
	return ""
}

func WSALookupServiceBegin(querySet *WSAQUERYSET, flags uint32, handle *windows.Handle) error {
	var qs = unsafe.Pointer(querySet)

	r, _, errNo := syscall.SyscallN(procWSALookupServiceBegin.Addr(), uintptr(qs), uintptr(flags), uintptr(unsafe.Pointer(handle)))
	if r == socket_error {
		return errnoErr(errNo)
	}

	return nil
}

func WSALookupServiceNext(handle windows.Handle, flags uint32, size *int32) (WSAQUERYSET, error) {
	var data = initializeWSAQUERYSET()
	r, _, errNo := syscall.SyscallN(procWSALookupServiceNext.Addr(), uintptr(handle), uintptr(flags), uintptr(unsafe.Pointer(size)), uintptr(unsafe.Pointer(data)))
	if r == socket_error {
		return WSAQUERYSET{}, errnoErr(errNo)
	}

	var ret WSAQUERYSET
	if data != nil {
		ret = *data
	}

	return ret, nil
}

func WSALookupServiceEnd(handle windows.Handle) error {
	r, _, errNo := syscall.SyscallN(procWSALookupServiceEnd.Addr(), uintptr(handle))
	if r == socket_error {
		return errnoErr(errNo)
	}
	return nil
}

func initializeWSAQUERYSET() *WSAQUERYSET {
	var serviceInstanceName = new(uint16)
	var serviceClassId = new(windows.GUID)
	var version = new(WSAVersion)
	var nsProviderId = new(windows.GUID)
	var comment = new(uint16)
	var context = new(uint16)
	var afpProtocols = new(AFProtocols)
	var queryString = new(uint16)
	var saBuffer = &AddrInfo{
		LocalAddr: SocketAddress{
			Sockaddr: &Sockaddr{},
		},
		RemoteAddr: SocketAddress{
			Sockaddr: &Sockaddr{},
		},
	}
	var blob = &BLOB{
		BlobData: new(byte),
	}
	return &WSAQUERYSET{
		ServiceInstanceName: serviceInstanceName,
		ServiceClassId:      serviceClassId,
		Version:             version,
		NSProviderId:        nsProviderId,
		Comment:             comment,
		Context:             context,
		AfpProtocols:        afpProtocols,
		QueryString:         queryString,
		SaBuffer:            saBuffer,
		Blob:                blob,
	}
}

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	return e
}
