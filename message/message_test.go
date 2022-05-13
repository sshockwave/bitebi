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

var tx1 Transaction = Transaction{
	Version: 2203,
	Tx_in: []txIn{
		{
			Previous_output: Outpoint{
				Hash: [32]byte{33,22,0,11},
				Index: 12,
			},
			script_bytes: 4,
			signature_script: []byte{22,1,1,4},
		},
		{
			Previous_output: Outpoint{
				Hash: [32]byte{8,2,6,3},
				Index: 7,
			},
			script_bytes: 5,
			signature_script: []byte{1,2,3,4,5},
		},
	},
	Tx_out: []txOut{
		{
			Value: 22555343,
			Pk_script: []byte{12,3,7,5},
		},
		{
			Value: -33,
			Pk_script: []byte{0,0,3,0,4},
		},
		{
			Value: 22,
			Pk_script: []byte{},
		},
	},
	Lock_time: 332,
}
func TestTxSerializing(t *testing.T) {
	b, err := utils.GetBytes(&tx1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var tx2 Transaction
	reader := utils.NewBufReader(bytes.NewBuffer(b))
	err = tx2.LoadBuffer(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(tx1, tx2) {
		t.Fatalf("Expected equal: %v = %v", tx1, tx2)
	}
}
