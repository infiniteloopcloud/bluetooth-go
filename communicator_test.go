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
		{
			in:  "2C:54:91:88:C9:E3",
			out: 0x2c549188c9e3,
		},
		{
			in:  "58:CF:0A:BB:28:FC",
			out: 0x58cf0abb28fc,
		},
	}
	for _, scenario := range scenarios {
		result, err := addressToUint64(scenario.in)
		if err != nil {
			t.Fatal(err)
		}
		if result != scenario.out {
			t.Errorf("Result should be %x, instead of %x", scenario.out, result)
		}
	}
}
