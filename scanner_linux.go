package bluetooth

import (
	"fmt"
	"os/exec"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

var _ Scanner = scanner{}

type scanner struct{}

const (
	ioctlSize = 4
	typHCI    = 72 // 'H'

	ogfLinkCtl                  = 0x01
	ocfRemoteNameReq            = 0x0019
	eventRemoteNameResponse     = 0x07
	remoteNameRequestSize       = 10
	eventRemoteNameResponseSize = 255

	evtRemoteNameReqComplete = 0x07
	evtCmdStatus             = 0x0f
	evtCmdComplete           = 0x0e
	evtLeMetaEvent           = 0x3e

	solHci       = 0
	hciFilterOpt = 2

	hciCommandPkt uint8 = 0x01

	hciCommandHdrSize = 3
	hciMaxEventSize   = 260
	hciEventHdrSize   = 2
)

var hciInquiry = ioR(typHCI, 240, ioctlSize)

func NewScanner() Scanner {
	return scanner{}
}

// ScanDeprecated is the old and deprecated implementation, I would like to keep it for a while as a fallback
func (scanner) ScanDeprecated() ([]Device, error) {
	out, err := exec.Command("hcitool", "scan").Output()
	if err != nil {
		return nil, fmt.Errorf("hcitool scan: %s", err.Error())
	}
	var devices []Device
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		for i, line := range lines {
			if i == 0 {
				continue
			}
			lineElems := strings.Split(line, string([]byte{9}))
			if len(lineElems) > 2 {
				devices = append(devices, Device{
					Name:       lineElems[2],
					MACAddress: lineElems[1],
				})
			}
		}
	}

	return devices, nil
}

func (scanner) Scan() ([]Device, error) {
	resp, err := inquiry()
	if err != nil {
		return nil, err
	}

	fd, err := openDevice()
	if err != nil {
		return nil, err
	}

	var devices []Device
	for i := 0; i < int(resp.numberOfResponses); i++ {
		name, err := readRemoteName(fd, resp.response[i])
		if err != nil {
			return nil, err
		}
		devices = append(devices, Device{
			Name:       string(name),
			MACAddress: byteArrayToAddress(resp.response[i].bluetoothDeviceAddress),
		})
	}

	return devices, nil
}

func inquiry() (hciInquiryRequest, error) {
	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.BTPROTO_HCI)
	if err != nil {
		return hciInquiryRequest{}, fmt.Errorf("inquiry: open unix socket - %s", err.Error())
	}

	req := hciInquiryRequest{
		length: 8,
	}
	req.lap[0] = 0x33
	req.lap[1] = 0x8b
	req.lap[2] = 0x9e

	if err = ioctl(fd, hciInquiry, unsafe.Pointer(&req)); err != nil {
		return hciInquiryRequest{}, fmt.Errorf("inquiry: hciInquiry ioctl call - %s", err.Error())
	}
	return req, nil
}

func openDevice() (int, error) {
	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.BTPROTO_HCI)
	if err != nil {
		return 0, fmt.Errorf("openDevice: open unix socket - %s", err.Error())
	}

	sock := unix.SockaddrHCI{
		Dev:     0,
		Channel: 0,
	}
	err = unix.Bind(fd, &sock)
	if err != nil {
		return 0, fmt.Errorf("openDevice: bind hci to unix socket - %s", err.Error())
	}
	return fd, nil
}

func readRemoteName(fd int, info inquiryInfo) ([]byte, error) {
	var response = remoteNameResponse{}
	var request = remoteNameRequest{
		bluetoothDeviceAddress: addressCopy(info.bluetoothDeviceAddress),
		scanRepMode:            info.scanRepMode,
		clockOffset:            info.clockOffset,
	}
	var hr = hciRequest{
		ogf:         ogfLinkCtl,
		ocf:         ocfRemoteNameReq,
		event:       eventRemoteNameResponse,
		request:     &request,
		requestLen:  remoteNameRequestSize,
		response:    &response,
		responseLen: eventRemoteNameResponseSize,
	}
	opcode := opcodePack(hr.ogf, hr.ocf)

	// TODO add better implementation for that filter
	var filter = hciFilter{
		typeMask: 16,
		eventMask: [2]uint32{
			0: 49280,
			1: 1073741824,
		},
		opcode: opcode,
	}
	err := setsockopt(fd, solHci, hciFilterOpt, unsafe.Pointer(&filter), unsafe.Sizeof(filter))
	if err != nil {
		return nil, fmt.Errorf("readRemoteName: call setsockopt with hciFilterOpt  - %s", err.Error())
	}

	err = sendCommand(fd, hr.ogf, hr.ocf, uint8(hr.requestLen), hr.request)
	if err != nil {
		return nil, fmt.Errorf("readRemoteName: call sendCommand - %s", err.Error())
	}

	err = pollSocket(fd)
	if err != nil {
		return nil, fmt.Errorf("readRemoteName: call pollSocket - %s", err.Error())
	}

	var try = 10
	for i := 0; i < try; i++ {
		name, ok, err := readEvent(fd)
		if err != nil {
			return nil, err
		}
		if ok {
			return name, nil
		}
	}

	return nil, nil
}

func readEvent(fd int) ([]byte, bool, error) {
	var buf = make([]byte, hciMaxEventSize)
	_, err := unix.Read(fd, buf)
	if err != nil {
		return nil, false, err
	}

	hdr := (*hciEventHdr)(unsafe.Pointer(&buf[1]))
	ptr := 1 + hciEventHdrSize

	switch hdr.event {
	case evtRemoteNameReqComplete:
		data := (*remoteNameResponse)(unsafe.Pointer(&buf[ptr]))
		return normalizeName(data.name), true, nil
	case evtCmdStatus:
		return nil, false, nil
	case evtCmdComplete:
		return nil, false, nil
	case evtLeMetaEvent:
		return nil, false, nil
	}

	return nil, false, nil
}

func pollSocket(fd int) error {
	timeout := 100000
	pollFd := unix.PollFd{
		Fd:     int32(fd),
		Events: unix.POLLIN,
	}
	_, err := unix.Poll([]unix.PollFd{pollFd}, timeout)
	return err
}

func sendCommand(fd int, ogf uint16, ocf uint16, reqLen uint8, req *remoteNameRequest) error {
	var commandType = hciCommandPkt
	var hc hciCommandHdr
	hc.opcode = opcodePack(ogf, ocf)
	hc.pLen = reqLen

	var iv = []unix.Iovec{
		{
			Base: (*byte)(unsafe.Pointer(&commandType)),
			Len:  1,
		},
		{
			Base: (*byte)(unsafe.Pointer(&hc)),
			Len:  hciCommandHdrSize,
		},
	}
	if reqLen > 0 {
		iv = append(iv, unix.Iovec{
			Base: (*byte)(unsafe.Pointer(req)),
			Len:  uint64(reqLen),
		})
	}

	_, err := writev(fd, iv)
	return err
}

func opcodePack(ogf, ocf uint16) uint16 {
	return (ocf & 0x03ff) | (ogf << 10)
}

func addressCopy(in [6]uint8) [6]uint8 {
	var out [6]uint8
	//nolint:gosimple
	for i := range in {
		out[i] = in[i]
	}
	return out
}

func normalizeName(in [248]uint8) []byte {
	var out []byte
	for i := range in {
		if in[i] != 0 {
			if in[i] < 32 || in[i] == 127 {
				out = append(out, '.')
			} else {
				out = append(out, in[i])
			}
		}
	}
	return out
}
