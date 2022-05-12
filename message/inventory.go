package message

import (
	"errors"

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
func (o *Inventory) PutBuffer(writer utils.BufWriter) (err error) {
    err = writer.WriteUint32(o.Type)
    if err != nil {
        return
    }
    err = writer.Write32Bytes(o.Hash)
    return
}

// Inventory
// https://developer.bitcoin.org/reference/p2p_networking.html#inv
type InvMsg struct {
    Inv []Inventory
}
const InvMaxItemCount = 50_000;
var invItemCountExceeded = errors.New("invItemCountExceeded")
func (ret *InvMsg) LoadBuffer(reader utils.BufReader) (err error) {
    var cnt uint64
    cnt, err = reader.ReadCompactUint()
    if err != nil {
        return
    }
    if cnt > InvMaxItemCount {
        return invItemCountExceeded
    }
    ret.Inv = make([]Inventory, cnt)
    for i := uint64(0); i < cnt; i++ {
        ret.Inv[i], err = NewInventory(reader)
        if err != nil {
            return
        }
    }
    return
}

func (m *InvMsg) PutBuffer(writer utils.BufWriter) (err error) {
    err = writer.WriteCompactUint(uint64(len(m.Inv)))
    if err != nil {
        return
    }
    for i := range m.Inv {
        err = m.Inv[i].PutBuffer(writer)
        if err != nil {
            return
        }
    }
    return
}
