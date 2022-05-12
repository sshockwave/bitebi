package message

import (
	"errors"
	"time"

	"github.com/sshockwave/bitebi/utils"
)

type Block struct {
	Version                    int32
	Previous_block_header_hash [32]byte
	Merkle_root_hash           [32]byte
	Time                       uint32
	NBits                      uint32
	Nonce                      uint32
}

func (b *Block) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt32(b.Version)
	if err != nil {
		return
	}
	err = writer.Write32Bytes(b.Previous_block_header_hash)
	if err != nil {
		return
	}
	err = writer.Write32Bytes(b.Merkle_root_hash)
	if err != nil {
		return
	}
	err = writer.WriteUint32(b.Time)
	if err != nil {
		return
	}
	err = writer.WriteUint32(b.NBits)
	if err != nil {
		return
	}
	err = writer.WriteUint32(b.Nonce)
	return
}

func (b *Block) LoadBuffer(reader utils.BufReader) (err error) {
	b.Version, err = reader.ReadInt32()
	if err != nil {
		return
	}
	b.Previous_block_header_hash, err = reader.Read32Bytes()
	if err != nil {
		return
	}
	b.Merkle_root_hash, err = reader.Read32Bytes()
	if err != nil {
		return
	}
	b.Time, err = reader.ReadUint32()
	if err != nil {
		return
	}
	b.NBits, err = reader.ReadUint32()
	if err != nil {
		return
	}
	b.Nonce, err = reader.ReadUint32()
	if err != nil {
		return
	}
	return
}

var blockHashNotValid = errors.New("blockHashNotValid")

func CreateBlock(version int32, previous_block_header_hash [32]byte, TS []Transaction, nBits uint32, nonce uint32) (Block, error) {
	var block Block
	block.Version = version
	block.Previous_block_header_hash = previous_block_header_hash
	block.Merkle_root_hash = MakeMerkleTree(TS)
	block.Time = uint32(time.Now().Unix())
	block.NBits = nBits
	block.Nonce = nonce

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

// https://developer.bitcoin.org/reference/block_chain.html#serialized-blocks
type SerializedBlock struct {
	Header     Block
	HeaderHash [32]byte
	Txns       []Transaction
}

func (b *SerializedBlock) PutBuffer(writer utils.BufWriter) (err error) {
	err = b.Header.PutBuffer(writer)
	if err != nil {
		return
	}
	err = writer.WriteCompactUint(uint64(len(b.Txns)))
	if err != nil {
		return
	}
	for i := range b.Txns {
		err = b.Txns[i].PutBuffer(writer)
		if err != nil {
			return
		}
	}
	return
}

func CreateSerialBlock(block Block, tx []Transaction) (sb SerializedBlock, err error) {
	sb.Header = block
	sb.Txns = tx
	sb.HeaderHash, err = utils.GetHash(&block)
	return
}

func (b *SerializedBlock) LoadBuffer(reader utils.BufReader) (err error) {
	err = b.Header.LoadBuffer(reader)
	if err != nil {
		return
	}
	b.HeaderHash, err = utils.GetHash(&b.Header)
	var cnt uint64
	cnt, err = reader.ReadCompactUint()
	if err != nil {
		return
	}
	b.Txns = make([]Transaction, cnt)
	for i := uint64(0); i < cnt; i++ {
		err = b.Txns[i].LoadBuffer(reader)
	}
	return
}
