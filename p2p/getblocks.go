package p2p

import (
    "bytes"
    "errors"
    "github.com/sshockwave/bitebi/utils"
)

// https://developer.bitcoin.org/reference/p2p_networking.html#getblocks
type GetBlocksMsg struct {
    version uint32
    // Hashes should be provided in reverse order of block height,
    // so highest-height hashes are listed first and lowest-height hashes are listed last.
    block_header_hashes [][32]byte
    // if all zero, request an "inv" message
    // otherwise its the last header hash being requested
    // not included in the array above
    stop_hash [32]byte
}
var maxSizeExceededError = errors.New("maxSizeExceededError")
func NewGetBlocksMsg(data []byte) (ret GetBlocksMsg, err error) {
    reader := utils.NewBufReader(bytes.NewReader(data))
    ret.version, err = reader.ReadUint32()
    if err != nil {
        return
    }
    hash_count, err := reader.ReadCompactUint()
    if err != nil {
        return
    }
    const MAX_SIZE uint64 = 0x02000000;
    if hash_count > MAX_SIZE {
        return ret, maxSizeExceededError
    }
    ret.block_header_hashes = make([][32]byte, hash_count)
    for i := uint64(0); i < hash_count; i++ {
        ret.block_header_hashes[i], err = reader.Read32Bytes()
        if err != nil {
            return
        }
    }
    ret.stop_hash, err = reader.Read32Bytes()
    return
}
