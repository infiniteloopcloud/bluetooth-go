package bluetooth

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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
	var flags uint32 = windows.LUP_CONTAINERS
	flags |= windows.LUP_RETURN_NAME
	flags |= windows.LUP_RETURN_ADDR

	var querySet windows.WSAQUERYSET
	querySet.NameSpace = windows.NS_BTH
	querySet.Size = uint32(unsafe.Sizeof(windows.WSAQUERYSET{}))

	var handle windows.Handle
	err := windows.WSALookupServiceBegin(&querySet, flags, &handle)
	if err != nil {
		if errors.Is(err, windows.WSASERVICE_NOT_FOUND) {
			return devices, nil
		}
		return devices, err
	}
	defer windows.WSALookupServiceEnd(handle)

	n := int32(unsafe.Sizeof(windows.WSAQUERYSET{}))
	buf := make([]byte, n)
itemsLoop:
	for {
		q := (*windows.WSAQUERYSET)(unsafe.Pointer(&buf[0]))
		err := windows.WSALookupServiceNext(handle, flags, &n, q)
		switch err {
		case windows.WSA_E_NO_MORE, windows.WSAENOMORE:
			// no more data available - break the loop
			break itemsLoop
		case windows.WSAEFAULT:
			// buffer is too small - reallocate and try again
			buf = make([]byte, n)
		case nil:
			// found a record - display the item and fetch next item
			var addr string
			for _, e := range q.SaBuffer.RemoteAddr.Sockaddr.Addr.Data {
				if e != 0 {
					addr += fmt.Sprintf("%x", e)
				}
			}
			devices = append(devices, Device{
				Name:       windows.UTF16PtrToString(q.ServiceInstanceName),
				MACAddress: parseRemoteSockaddr(addr),
			})

		default:
			return devices, err
		}
	}

	return devices, nil
}

func parseRemoteSockaddr(addr string) string {
	var spaced = ""
	addr = strings.ReplaceAll(addr, "-", "")
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
