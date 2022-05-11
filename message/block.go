package message

import (
	"crypto/sha256"
	"fmt"
)

type Block struct {
	version                    int32
	previous_block_header_hash [32]byte
	merkle_root_hash           [32]byte
	time                       uint32
	nBits                      uint32
	nonce                      uint32
}

func WriteBlock(previous_block_header_hash [32]byte, TS []Transaction, nonce uint32) Block {
	var block Block
	block.previous_block_header_hash = previous_block_header_hash
	block.merkle_root_hash = makeMerkleTree(TS)

	block.nonce = nonce
}

func hash(src []byte) [32]byte {
	res := sha256.Sum256(src)
	return res
}

/*
func string2byteslice(s string) [32]uint8 {
	var array [64]uint8
	var output [32]byte
	for i := 0; i < 64; i++ {
		if s[i] >= 48 && s[i] <= 57 {
			array[i] = s[i] - 48
		} else if s[i] >= 97 && s[i] <= 102 {
			array[i] = s[i] - 87
		}
	}

	for i := 0; i < 32; i++ {
		output[i] = 16*array[2*i] + array[2*i+1]
	}
	return output
}*/

func main() {
	src := []byte("1234")
	h := sha256.New()
	h.Write(src)
	res := h.Sum(nil)
	fmt.Printf("%T\n", res)
	fmt.Println(res[31])
}
