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
