package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type txOut struct {
	value     int64
	pk_script []byte
}

func NewtxOut(value int64, pk_script_bytes string) txOut {
	var to txOut
	to.value = value
	to.pk_script = []byte(pk_script_bytes)
	return to
}

func (t *txOut) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt64(t.value)
	if err != nil {
		return err
	}
	err = writer.WriteBytes(t.pk_script)
	return
}
