package message

import (
	"bytes"
	"errors"

	"github.com/sshockwave/bitebi/utils"
)

type MerkleTree struct {
	Hash       [32]byte
	Lson, Rson *MerkleTree
	IsLeaf     bool
	// "Used" means that this transaction is included in this Merkle Tree
	// Unused nodes appear in the tree because they are needed for hash calculation
	// This variable makes sense only if IsLeaf is true
	Used bool
}

type merkleTreeBuilder struct {
	Hashes           [][32]byte
	Flags            []byte
	PosHash, PosFlag int
}

var childrenSameHashError = errors.New("childrenSameHashError")
var MerkleBlockNotEnoughFlags = errors.New("MerkleBlockNotEnoughFlags")
var MerkleBlockNotEnoughHash = errors.New("MerkleBlockNotEnoughHash")

// https://developer.bitcoin.org/reference/p2p_networking.html#parsing-a-merkleblock-message
func (b *merkleTreeBuilder) BuildMerkleTree(depth int) (node *MerkleTree, err error) {
	if b.PosFlag == len(b.Flags)*8 {
		return nil, MerkleBlockNotEnoughFlags
	}
	flag := ((b.Flags[b.PosFlag>>3] >> (b.PosFlag & 7)) & 1) == 1
	b.PosFlag += 1
	node = new(MerkleTree)
	if !(depth > 0 && flag) {
		// the only case that hash is not needed:
		// non root and both children are accessed
		if b.PosHash == len(b.Hashes) {
			return nil, MerkleBlockNotEnoughHash
		}
		node.Hash = b.Hashes[b.PosHash]
		b.PosHash += 1
	}
	if depth == 0 {
		node.IsLeaf = true
		node.Used = flag
	} else {
		if flag {
			node.Lson, err = b.BuildMerkleTree(depth - 1)
			if err != nil {
				return
			}
			node.Rson, err = b.BuildMerkleTree(depth - 1)
			if err != nil {
				return
			}
			// CVE-2012-2459
			if node.Lson.Hash == node.Rson.Hash {
				return node, childrenSameHashError
			}
			node.Hash = utils.Sha256Twice(bytes.Join([][]byte{node.Lson.Hash[:], node.Rson.Hash[:]}, []byte{}))
		} else {
			if b.PosHash == len(b.Hashes) {
				return nil, MerkleBlockNotEnoughHash
			}
			node.Hash = b.Hashes[b.PosHash]
			b.PosHash += 1
		}
	}
	return
}

type MerkleBlockMsg struct {
	// TODO: Block header
	TxCount uint32
	root    *MerkleTree
}

var MerkleBlockTooManyHashes = errors.New("MerkleBlockTooManyHashes")
var MerkleBlockTooManyFlags = errors.New("MerkleBlockTooManyFlags")

func NewMerkleBlockMsg(reader utils.BufReader) (ret MerkleBlockMsg, err error) {
	// TODO: parse block header, 80 bytes
	ret.TxCount, err = reader.ReadUint32()
	if err != nil {
		return
	}
	var builder merkleTreeBuilder
	var hash_cnt uint64
	hash_cnt, err = reader.ReadCompactUint()
	if err != nil {
		return
	}
	builder.Hashes = make([][32]byte, hash_cnt)
	for i := uint64(0); i < hash_cnt; i++ {
		builder.Hashes[i], err = reader.Read32Bytes()
		if err != nil {
			return
		}
	}
	var flag_cnt uint64
	flag_cnt, err = reader.ReadCompactUint()
	if err != nil {
		return
	}
	builder.Flags = make([]byte, flag_cnt)
	builder.Flags, err = reader.ReadBytes(int(flag_cnt))
	if err != nil {
		return
	}
	// node.root = builder.BuildMerkleTree(depth)
	if uint64(builder.PosHash) < hash_cnt {
		return ret, MerkleBlockTooManyHashes
	}
	if builder.PosFlag < int(flag_cnt*8)-8 {
		// Last flag byte is not used
		return ret, MerkleBlockTooManyFlags
	}
	// TODO: check if root hash matches the block merkle root
	return
}
