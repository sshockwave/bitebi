package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type TxIn struct {
	Previous_output  Outpoint
	Script_bytes     uint64
	Signature_script []byte
	Sequence         uint32
}

func (t *TxIn) PutBuffer(writer utils.BufWriter) (err error) {
	err = t.Previous_output.PutBuffer(writer)
	if err != nil {
		return
	}
	err = writer.WriteCompactUint(t.Script_bytes)
	if err != nil {
		return
	}
	err = writer.WriteBytes(t.Signature_script)
	if err != nil {
		return
	}
	err = writer.WriteUint32(t.Sequence)
	return
}

func NewtxIn(previous_output Outpoint, script_bytes uint64, signature_script []byte, sequence uint32) TxIn {
	var ti TxIn
	ti.Previous_output = previous_output
	ti.Script_bytes = script_bytes
	ti.Signature_script = signature_script
	ti.Sequence = sequence
	return ti
}

type Outpoint struct {
	Hash  [32]byte
	Index uint32
}

func (o *Outpoint) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.Write32Bytes(o.Hash)
	if err != nil {
		return
	}
	err = writer.WriteUint32(o.Index)
	return
}

func NewOutPoint(hash [32]byte, index uint32) Outpoint {
	var op Outpoint
	op.Hash = hash
	op.Index = index
	return op
}

func (data *TxIn) LoadBuffer(reader utils.BufReader) (err error) {
	err = data.Previous_output.LoadBuffer(reader)
	if err != nil {
		return
	}
	data.Script_bytes, err = reader.ReadCompactUint()
	if err != nil {
		return
	}
	data.Signature_script, err = reader.ReadBytes(int(data.Script_bytes))
	if err != nil {
		return
	}
	data.Sequence, err = reader.ReadUint32()
	return
}

func (o *Outpoint) LoadBuffer(reader utils.BufReader) (err error) {
	o.Hash, err = reader.Read32Bytes()
	if err != nil {
		return
	}
	o.Index, err = reader.ReadUint32()
	return
}
