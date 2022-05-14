package utils

import (
	"reflect"
	"testing"
)

// https://developer.bitcoin.org/reference/p2p_networking.html#message-headers
var emptyHash = []byte{0x5d, 0xf6, 0xe0, 0xe2}
func TestBitcoinEmptyHash(t *testing.T) {
	hash := Sha256Twice([]byte{})
	if !reflect.DeepEqual(emptyHash, hash[:4]) {
		t.Fatalf("Expect empty hash to be %v, %v found", emptyHash, hash)
	}
}

func TestHashNBits(t *testing.T) {
	hash := [32]byte{0,0,0,0,0,0x3f,0x22}
	res := HasValidHash(hash, 0x07223f00)
	if res != true {
		t.Fatal()
	}
	res = HasValidHash(hash, 0x072233ff)
	if res != false {
		t.Fatal()
	}
	res = HasValidHash(hash, 0x06223f00)
	if res != false {
		t.Fatal()
	}
	res = HasValidHash(hash, 0x08223f00)
	if res != true {
		t.Fatal()
	}
}
