package message

import (
    "github.com/sshockwave/bitebi/utils"
)


// https://developer.bitcoin.org/reference/p2p_networking.html#data-messages
const (
    MSG_TX = 1
    MSG_BLOCK = 2
    MSG_FILTERED_BLOCK = 3
    MSG_CMPCT_BLOCK= 4
    MSG_WITNESS_FLAG = 1 << 30
    MSG_WITNESS_TX = MSG_TX | MSG_WITNESS_FLAG
    MSG_WITNESS_BLOCK = MSG_BLOCK | MSG_WITNESS_FLAG
    MSG_FILTERED_WITNESS_BLOCK = MSG_FILTERED_BLOCK | MSG_WITNESS_FLAG
)

type Inventory struct {
    Type uint32
    Hash [32]byte
}
func NewInventory(reader utils.BufReader) (ret Inventory, err error) {
    ret.Type, err = reader.ReadUint32()
    if err != nil {
        return
    }
    ret.Hash, err = reader.Read32Bytes()
    return
}

// Inventory
// https://developer.bitcoin.org/reference/p2p_networking.html#inv
type InvMsg struct {
    inv []Inventory
}
func NewInvMsg(reader utils.BufReader) (ret InvMsg, err error) {
    var cnt uint64
    cnt, err = reader.ReadCompactUint()
    if err != nil {
        return
    }
    ret.inv = make([]Inventory, cnt)
    for i := uint64(0); i < cnt; i++ {
        ret.inv[i], err = NewInventory(reader)
        if err != nil {
            return
        }
    }
    return
}
