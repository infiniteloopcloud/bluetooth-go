package bluetooth

import (
	"fmt"
	"testing"
)

func TestHCITool_Scan(t *testing.T) {
	devices, _ := NewScanner().Scan()
	fmt.Printf("%+v\n", devices)
}

// NOTE: this test only for windows
//func TestParseRemoteSockaddr(t *testing.T) {
//	var scenarios = []struct {
//		in  string
//		out string
//	}{
//		{
//			in:  "d2cd7f2d8154",
//			out: "54:81:2D:7F:CD:D2",
//		},
//		{
//			in:  "ca7521e01a74",
//			out: "74:1A:E0:21:75:CA",
//		},
//	}
//	for _, scenario := range scenarios {
//		out := parseRemoteSockaddr(scenario.in)
//		if scenario.out != out {
//			t.Errorf("Should be %s, instead of %s", scenario.out, out)
//		}
//	}
//
//}
