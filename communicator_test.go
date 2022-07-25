package bluetooth

import "testing"

func TestAddressToUint64(t *testing.T) {
	scenarios := []struct {
		in  string
		out uint64
	}{
		{
			in:  "54:81:2D:7E:CD:D2",
			out: 0x54812d7ecdd2,
		},
	}
	for _, scenario := range scenarios {
		result := addressToUint64(scenario.in)
		if result != scenario.out {
			t.Error("invalid")
		}
	}
}
