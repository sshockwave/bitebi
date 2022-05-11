package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type txIn struct {
	Previous_output  outpoint
	script_bytes     uint64
	signature_script []byte
	sequence         uint32
}

func (t *txIn) PutBuffer(writer utils.BufWriter) (err error) {
	err = t.Previous_output.PutBuffer(writer)
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

func NewtxIn(previous_output outpoint, script_bytes uint64, signature_script []byte, sequence uint32) txIn {
	var ti txIn
	ti.Previous_output = previous_output
	ti.script_bytes = script_bytes
	ti.signature_script = signature_script
	ti.sequence = sequence
	return ti
}

type outpoint struct {
	Hash  [32]byte
	Index uint32
}

func (o *outpoint) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.Write32Bytes(o.Hash)
	if err != nil {
		return
	}
	err = writer.WriteUint32(o.Index)
	return
}

func Newoutpoint(hash [32]byte, index uint32) outpoint {
	var op outpoint
	op.Hash = hash
	op.Index = index
	return op
}
