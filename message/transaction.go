package message

type Transaction struct {
	version   int32
	tx_in     txIn
	tx_out    txOut
	lock_time uint32
}

func hashTransaction(ts Transaction) [32]byte {

}

func makeMerkleTree(TS []Transaction) [32]byte {
	if len(TS) == 1 {
		return hashTransaction(TS[0])
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
