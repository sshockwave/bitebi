package message

import (
	"crypto/sha256"

	"github.com/sshockwave/bitebi/utils"
)

type Transaction struct {
	Version      int32
	Tx_in_count  uint64
	Tx_in        []txIn
	Tx_out_count uint64
	Tx_out       []txOut
	Lock_time    uint32
}

func (t *Transaction) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt32(t.Version)
	if err != nil {
		return
	}
	err = t.Tx_in.PutBuffer(writer)
	if err != nil {
		return
	}
	err = t.Tx_out.PutBuffer(writer)
	if err != nil {
		return
	}
	err = writer.WriteUint32(t.Lock_time)
	return
}

func CreateTransaction(version int32, tx_in_count uint64, tx_in []txIn, tx_out_count uint64, tx_out []txOut, lock_time uint32) Transaction {
	var ts Transaction
	ts.Version = version
	ts.Tx_in_count = tx_in_count
	ts.Tx_in = tx_in
	ts.Tx_out_count = tx_out_count
	ts.Tx_out = tx_out
	ts.Lock_time = lock_time
	return ts
}

func makeMerkleTree(TS []Transaction) [32]byte {
	if len(TS) == 1 {
		hash, _ := utils.GetHash(&TS[0])
		return hash
	} else {
		var m int
		m = len(TS) / 2
		TS1 := TS[:m+1]
		TS2 := TS[m+1:]
		hash1 := makeMerkleTree(TS1)
		hash2 := makeMerkleTree(TS2)
		var src [64]byte
		for i := 0; i < 32; i++ {
			src[i] = hash1[i]
			src[32+i] = hash2[i]
		}
		res := sha256.Sum256(src[:])
		res = sha256.Sum256(res[:])
		return res
	}
}

func (tx *Transaction) LoadBuffer(reader utils.BufReader) (err error) {
	tx.Version, err = reader.ReadInt32()
	if err != nil {
		return
	}
	tx.Tx_in_count, err = reader.ReadCompactUint()
	tx.Tx_in = make([]txIn, tx.Tx_in_count)
	for i := uint64(0); i < tx.Tx_in_count; i++ {
		err = tx.Tx_in[i].LoadBuffer(reader)
		if err != nil {
			return
		}
	}
	tx.Tx_out_count, err = reader.ReadCompactUint()
	tx.Tx_out = make([]txOut, tx.Tx_out_count)
	for i := uint64(0); i < tx.Tx_out_count; i++ {
		err = tx.Tx_out[i].LoadBuffer(reader)
		if err != nil {
			return
		}
	}
	return
}
