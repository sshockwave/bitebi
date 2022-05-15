package message

import (
	"bytes"

	"github.com/sshockwave/bitebi/utils"
)

type Transaction struct {
	Version   int32
	Tx_in     []TxIn
	Tx_out    []TxOut
	Lock_time uint32
}

func (t *Transaction) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt32(t.Version)
	if err != nil {
		return
	}
	err = writer.WriteCompactUint(uint64(len(t.Tx_in)))
	if err != nil {
		return
	}
	for i := range t.Tx_in {
		err = t.Tx_in[i].PutBuffer(writer)
		if err != nil {
			return
		}
	}
	err = writer.WriteCompactUint(uint64(len(t.Tx_out)))
	if err != nil {
		return
	}
	for i := range t.Tx_out {
		err = t.Tx_out[i].PutBuffer(writer)
		if err != nil {
			return
		}
	}
	if err != nil {
		return
	}
	err = writer.WriteUint32(t.Lock_time)
	return
}

func CreateTransaction(version int32, tx_in []TxIn, tx_out []TxOut, lock_time uint32) Transaction {
	var ts Transaction
	ts.Version = version
	ts.Tx_in = tx_in
	ts.Tx_out = tx_out
	ts.Lock_time = lock_time
	return ts
}

const HashL = 32

func MakeMerkleTree(TS []Transaction) (res [32]byte) {
	var tmp [HashL]byte
	hashes := make([]byte, len(TS)*HashL)
	for i := range TS {
		tmp, _ = utils.GetHash(&TS[i])
		copy(hashes[i*HashL:(i+1)*HashL], tmp[:])
	}
	return MakeMerkleTreeFromHashes(hashes)
}
func MakeMerkleTreeFromHashes(hashes []byte) (res [32]byte) {
	for n := len(hashes) / HashL; n > 1; n = (n + 1) / 2 {
		for i := 0; i < n/2; i++ {
			res = utils.Sha256Twice(hashes[2*i*HashL : (2*i+2)*HashL])
			copy(hashes[i*HashL:(i+1)*HashL], res[:])
		}
		if n%2 == 1 {
			p := hashes[(n-1)*HashL : n*HashL]
			res = utils.Sha256Twice(bytes.Join([][]byte{p, p}, []byte{}))
			copy(hashes[(n-1)/2*HashL:(n+1)/2*HashL], res[:])
		}
	}
	copy(res[:], hashes[:HashL])
	return
}

func (tx *Transaction) LoadBuffer(reader utils.BufReader) (err error) {
	tx.Version, err = reader.ReadInt32()
	if err != nil {
		return
	}
	var cnt uint64
	cnt, err = reader.ReadCompactUint()
	tx.Tx_in = make([]TxIn, cnt)
	for i := uint64(0); i < cnt; i++ {
		err = tx.Tx_in[i].LoadBuffer(reader)
		if err != nil {
			return
		}
	}
	cnt, err = reader.ReadCompactUint()
	tx.Tx_out = make([]TxOut, cnt)
	for i := uint64(0); i < cnt; i++ {
		err = tx.Tx_out[i].LoadBuffer(reader)
		if err != nil {
			return
		}
	}
	tx.Lock_time, err = reader.ReadUint32()
	return
}
