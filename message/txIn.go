package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type txIn struct {
	Previous_output  Outpoint
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

func NewtxIn(previous_output Outpoint, script_bytes uint64, signature_script []byte, sequence uint32) txIn {
	var ti txIn
	ti.Previous_output = previous_output
	ti.script_bytes = script_bytes
	ti.signature_script = signature_script
	ti.sequence = sequence
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

func (data *txIn) LoadBuffer(reader utils.BufReader) (err error) {
	err = data.Previous_output.LoadBuffer(reader)
	if err != nil {
		return
	}
	data.script_bytes, err = reader.ReadUint64()
	if err != nil {
		return
	}
	data.signature_script, err = reader.ReadBytes(int(data.script_bytes))
	if err != nil {
		return
	}
	data.sequence, err = reader.ReadUint32()
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
