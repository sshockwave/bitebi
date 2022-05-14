package utils

import (
	"crypto/sha256"
)

func Sha256Twice(data []byte) [32]byte {
    chksum1 := sha256.Sum256(data)
    return sha256.Sum256(chksum1[:])
}

func GetHash(data BinaryWritable) (hash [32]byte, err error) {
	var bytes []byte
	bytes, err = GetBytes(data)
	if err != nil {
		return
	}
	hash = Sha256Twice(bytes)
	return
}

// Specification in
// https://developer.bitcoin.org/reference/block_chain.html
// 0x18     1bc330
//   ^exp   ^significand
// assumes no negative nbits
func HasValidHash(hash [32]byte, nBits uint32) bool {
	exp := nBits >> 24
	if exp < 3 {
		return false
	}
	for _, v := range hash[exp:] {
		if v != 0 {
			return false
		}
	}
	target := nBits & ((1 << 24) - 1)
	value := uint32(hash[exp - 1]) << 16
	value += uint32(hash[exp - 2]) << 8
	value += uint32(hash[exp - 3])
	if value != target {
		return value < target
	}
	for _, v := range hash[:exp - 3] {
		if v != 0 {
			return false
		}
	}
	return true
}
