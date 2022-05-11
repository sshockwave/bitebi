package message

import (
    "errors"
    "bytes"
    "github.com/sshockwave/bitebi/utils"
)


type MerkleTree struct {
    Hash [32]byte
    Lson, Rson *MerkleTree
    IsLeaf bool
    // "Used" means that this transaction is included in this Merkle Tree
    // Unused nodes appear in the tree because they are needed for hash calculation
    // This variable makes sense only if IsLeaf is true
    Used bool
}

type MerkleBlockMsg struct {
    // TODO: Block header
    TxCount uint32
    HashCount uint64
    Hashes [][32]byte
    FlagByteCount uint64
    Flags []byte
}
var childrenSameHashError = errors.New("childrenSameHashError")
// https://developer.bitcoin.org/reference/p2p_networking.html#parsing-a-merkleblock-message
func (b *MerkleBlockMsg) BuildMerkleTree(pos_flag *int, pos_hash *int, depth int) (node *MerkleTree, err error) {
    flag := ((b.Flags[*pos_flag >> 3] >> (*pos_flag & 7)) & 1) == 1
    *pos_flag += 1
    node = new(MerkleTree)
    if depth == 0 {
        node.IsLeaf = true
        node.Hash = b.Hashes[*pos_hash]
        *pos_hash += 1
        node.Used = flag
    } else {
        if flag {
            node.Lson, err = b.BuildMerkleTree(pos_flag, pos_hash, depth - 1)
            if err != nil {
                return
            }
            node.Rson, err = b.BuildMerkleTree(pos_flag, pos_hash, depth - 1)
            if err != nil {
                return
            }
            // CVE-2012-2459
            if node.Lson.Hash == node.Rson.Hash {
                return node, childrenSameHashError
            }
            node.Hash = utils.Sha256Twice(bytes.Join([][]byte{node.Lson.Hash[:], node.Rson.Hash[:]}, []byte{}))
        } else {
            node.Hash = b.Hashes[*pos_hash]
            *pos_hash += 1
        }
    }
    return
}
func NewMerkleBlockMsg(reader utils.BufReader) (ret MerkleBlockMsg, err error){
    // TODO: parse block header, 80 bytes
    ret.TxCount, err = reader.ReadUint32()
    if err != nil {
        return
    }
    ret.HashCount, err = reader.ReadCompactUint()
    if err != nil {
        return
    }
    ret.Hashes = make([][32]byte, ret.HashCount)
    for i := uint64(0); i < ret.HashCount; i++ {
        ret.Hashes[i], err = reader.Read32Bytes()
        if err != nil {
            return
        }
    }
    ret.FlagByteCount, err = reader.ReadCompactUint()
    if err != nil {
        return
    }
    ret.Flags = make([]byte, ret.FlagByteCount)
    ret.Flags, err = reader.ReadBytes(int(ret.FlagByteCount))
    if err != nil {
        return
    }
    return
}
