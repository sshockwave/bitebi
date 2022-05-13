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
