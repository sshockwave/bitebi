package message

import (
	"github.com/sshockwave/bitebi/utils"
	"time"
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

func (b *Block) GetHash() (hash [32]byte, err error) {
	writer := utils.NewBufWriter()
	err = b.PutBuffer(writer)
	if err != nil {
		return
	}
	hash = utils.Sha256Twice(writer.Collect())
	return
}

func CreateBlock(version int32, previous_block_header_hash [32]byte, TS []Transaction, nBits uint32, nonce uint32) Block {
	var block Block
	block.version = version
	block.previous_block_header_hash = previous_block_header_hash
	block.merkle_root_hash = makeMerkleTree(TS)
	block.time = uint32(time.Now().Unix())
	block.nBits = nBits
	block.nonce = nonce

	if block.GetHash()

	return block
}

/*func hash(src []byte) [32]byte {
	res := sha256.Sum256(src)
	return res
}*/
