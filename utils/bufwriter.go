package utils

import (
	"bytes"
	"encoding/binary"
)


type BufWriter struct {
    out *bytes.Buffer
}

func NewBufWriter() (b BufWriter) {
    b.out = new(bytes.Buffer)
    return
}

func (b *BufWriter) Collect() [] byte {
    return b.out.Bytes()
}

func (b *BufWriter) WriteBytes(p []byte) (err error) {
    _, err = b.out.Write(p)
    return
}

func (b *BufWriter) Write32Bytes(p [32]byte) (err error) {
    return b.WriteBytes(p[:])
}

func (b *BufWriter) WriteUint8(v uint8) (err error) {
    return b.out.WriteByte(v)
}

func (b *BufWriter) WriteUint16(v uint16) (err error) {
    res := make([]byte, 2)
    binary.LittleEndian.PutUint16(res, v)
    return b.WriteBytes(res)
}

func (b *BufWriter) WriteUint32(v uint32) (err error) {
    res := make([]byte, 4)
    binary.LittleEndian.PutUint32(res, v)
    return b.WriteBytes(res)
}

func (b *BufWriter) WriteUint64(v uint64) (err error) {
    res := make([]byte, 8)
    binary.LittleEndian.PutUint64(res, v)
    return b.WriteBytes(res)
}

func (b *BufWriter) WriteCompactUint(v uint64) (err error) {
    if v <= 252 {
        return b.WriteUint8(uint8(v))
    }
    if v < 0x10000 {
        err = b.WriteUint8(253)
        if err != nil {
            return err
        }
        return b.WriteUint16(uint16(v))
    }
    if v < 0x100000000 {
        err = b.WriteUint8(254)
        if err != nil {
            return err
        }
        return b.WriteUint32(uint32(v))
    }
    err = b.WriteUint8(255)
    if err != nil {
        return err
    }
    return b.WriteUint64(v)
}

func (b *BufWriter) WriteInt32(v int32) (err error) {
    return binary.Write(b.out, binary.LittleEndian, v)
}

func (b *BufWriter) WriteInt64(v int64) (err error) {
    return binary.Write(b.out, binary.LittleEndian, v)
}

type BinaryWritable interface {
    PutBuffer(BufWriter) (err error)
}

func GetBytes(data BinaryWritable) (oput []byte, err error) {
	writer := NewBufWriter()
	err = data.PutBuffer(writer)
	if err != nil {
		return
	}
	oput = writer.Collect()
    return
}
