package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type txIn struct {
	previous_output  outpoint
	script_bytes     uint64
	signature_script []byte
	sequence         uint32
}

func (t *txIn) PutBuffer(writer utils.BufWriter) (err error) {
	err = t.previous_output.PutBuffer(writer)
	if err != nil {
		return
	}
	err = writer.WriteCompactUint(t.script_bytes)
	if err != nil {
		return
	}
	err = writer.WriteBytes(t.signature_script)
	if err != nil {
		return
	}
	err = writer.WriteUint32(t.sequence)
	return
}

func makeSlice(tx txIn) struct {
}

type outpoint struct {
	hash  [32]byte
	index uint32
}

func (o *outpoint) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.Write32Bytes(o.hash)
	if err != nil {
		return
	}
	err = writer.WriteUint32(o.index)
	return
}
