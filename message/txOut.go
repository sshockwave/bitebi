package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type txOut struct {
	Value     int64
	Pk_script []byte
}

func NewtxOut(value int64, pk_script_bytes string) txOut {
	var to txOut
	to.Value = value
	to.Pk_script = []byte(pk_script_bytes)
	return to
}

func (t *txOut) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt64(t.Value)
	if err != nil {
		return err
	}
	err = writer.WriteCompactUint(uint64(len(t.Pk_script)))
	if err != nil {
		return err
	}
	err = writer.WriteBytes(t.Pk_script)
	return
}

func (t *txOut) LoadBuffer(reader utils.BufReader) (err error) {
	t.Value, err = reader.ReadInt64()
	if err != nil {
		return err
	}
	var byte_cnt uint64
	byte_cnt, err = reader.ReadCompactUint()
	if err != nil {
		return err
	}
	t.Pk_script, err = reader.ReadBytes(int(byte_cnt))
	return
}
