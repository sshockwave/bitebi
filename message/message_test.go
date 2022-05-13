package message

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/sshockwave/bitebi/utils"
)

func TestAddr(t *testing.T) {
	addrs := AddrMsg{Addrs: []NetworkIPAddress{
		{
			time: 123456,
			services: 434432,
			Ipv6: [16]byte{0x7f,0x00,0x00,0,0,3,4,2},
			Port: 8333,
		},
		{
			time: 4433,
			services: 9929292,
			Ipv6: [16]byte{0x00,0x00,0x00,0,0,3,4,0x33},
			Port: 18333,
		},
	}}
	raw_data, err := utils.GetBytes(&addrs)
	if len(raw_data) == 0 || err != nil {
		t.Fatalf("Should have converted to bytes")
	}
	var new_addrs AddrMsg
	reader := utils.NewBufReader(bytes.NewBuffer(raw_data))
	err = new_addrs.LoadBuffer(reader)
	if err != nil {
		t.Fatal()
	}
	if !reflect.DeepEqual(addrs, new_addrs) {
		t.Fatalf("Expect equal: %v = %v", addrs, new_addrs)
	}
}
