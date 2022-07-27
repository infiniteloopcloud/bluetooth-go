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
	out, err := exec.Command("blueutil", "--inquiry").Output()
	if err != nil {
		return nil, fmt.Errorf("blueutil inquiry: %s", err.Error())
	}
	var devices []Device
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		for _, line := range lines {
			fmt.Println(line)
			parts := strings.Split(line, ",")
			var device = Device{}
			for _, elem := range parts {
				if strings.Contains(elem, "address:") {
					device.MACAddress = strings.ReplaceAll(elem, "address: ", "")
				}
				if strings.Contains(elem, "name:") {
					device.Name = strings.ReplaceAll(elem, "name: ", "")
					device.Name = strings.ReplaceAll(device.Name, `"`, "")
				}
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}
