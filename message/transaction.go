package message

import "github.com/sshockwave/bitebi/utils"

type Transaction struct {
	version   int32
	tx_in     txIn
	tx_out    txOut
	lock_time uint32
}

func (t *Transaction) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt32(t.version)
	if err != nil {
		return
	}
	err = t.tx_in.PutBuffer(writer)
	if err != nil {
		return
	}
	err = t.tx_out.PutBuffer(writer)
	if err != nil {
		return
	}
	err = writer.WriteUint32(t.lock_time)
	return
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
		res := hash(src[:])
		res = hash(res[:])
		return res
	}
}
