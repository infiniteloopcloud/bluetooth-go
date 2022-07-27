package bluetooth

import (
	"fmt"
	"os/exec"
	"strings"
)

var _ Scanner = scanner{}

type scanner struct{}

func NewScanner() Scanner {
	return scanner{}
}

func (scanner) Scan() ([]Device, error) {
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
