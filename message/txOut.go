package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type txOut struct {
	value     int64
	pk_script string
}

func (t *txOut) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteInt64(t.value)
	if err != nil {
		return err
	}
	err = writer.WriteBytes([]byte(t.pk_script))
	return
}
