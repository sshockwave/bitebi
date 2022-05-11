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
	err = writer.WriteBytes(t.Pk_script)
	return
}
