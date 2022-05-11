package message

import (
	"errors"
	"time"

	"github.com/sshockwave/bitebi/utils"
)

type Block struct {
	version                    int32
	previous_block_header_hash [32]byte
	merkle_root_hash           [32]byte
	time                       uint32
	nBits                      uint32
	nonce                      uint32
}

func (b *Block) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt32(b.version)
	if err != nil {
		return
	}
	err = writer.Write32Bytes(b.previous_block_header_hash)
	if err != nil {
		return
	}
	err = writer.Write32Bytes(b.merkle_root_hash)
	if err != nil {
		return
	}
	err = writer.WriteUint32(b.time)
	if err != nil {
		return
	}
	err = writer.WriteUint32(b.nBits)
	if err != nil {
		return
	}
	err = writer.WriteUint32(b.nonce)
	return
}

var blockHashNotValid = errors.New("blockHashNotValid")

func CreateBlock(version int32, previous_block_header_hash [32]byte, TS []Transaction, nBits uint32, nonce uint32) (Block, error) {
	var block Block
	block.version = version
	block.previous_block_header_hash = previous_block_header_hash
	block.merkle_root_hash = makeMerkleTree(TS)
	block.time = uint32(time.Now().Unix())
	block.nBits = nBits
	block.nonce = nonce

	hash, _ := utils.GetHash(&block)
	valid := false
	for i := 0; i < int(nBits); i++ {
		if hash[i/8] <= 255>>(i%8+1) {
			valid = true
		}
	}

	if valid == true {
		return block, nil
	} else {
		return block, blockHashNotValid
	}
}
