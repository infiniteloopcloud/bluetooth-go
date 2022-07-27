package bluetooth

import (
	"fmt"
	"testing"
)

func TestHCITool_Scan(t *testing.T) {
	devices, _ := NewScanner().Scan()
	fmt.Printf("%+v\n", devices)
}
