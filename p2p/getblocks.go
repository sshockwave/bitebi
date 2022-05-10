package p2p

import (
    "io"
    "errors"
    "github.com/sshockwave/bitebi/utils"
)

// https://developer.bitcoin.org/reference/p2p_networking.html#getblocks
type GetBlocksMsg struct {
    Version uint32
    // Hashes should be provided in reverse order of block height,
    // so highest-height hashes are listed first and lowest-height hashes are listed last.
    BlockHeaderHashes [][32]byte
    // if all zero, request an "inv" message
    // otherwise its the last header hash being requested
    // not included in the array above
    StopHash [32]byte
}
var maxSizeExceededError = errors.New("maxSizeExceededError")
func NewGetBlocksMsg(_reader io.Reader) (ret GetBlocksMsg, err error) {
    reader := utils.NewBufReader(_reader)
    ret.Version, err = reader.ReadUint32()
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
    ret.BlockHeaderHashes = make([][32]byte, hash_count)
    for i := uint64(0); i < hash_count; i++ {
        ret.BlockHeaderHashes[i], err = reader.Read32Bytes()
        if err != nil {
            return
        }
    }
    ret.StopHash, err = reader.Read32Bytes()
    return
}
