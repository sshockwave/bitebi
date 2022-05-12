package utils

import (
	"encoding/binary"
	"errors"
	"io"
)

type BufReader struct {
    src io.Reader
}

func NewBufReader(src io.Reader) BufReader {
    return BufReader{src}
}

func (b *BufReader) ReadBytes(n int) (data []byte, err error){
    data = make([]byte, n)
    n, err = io.ReadFull(b.src, data)
    return data, err
}

func (b *BufReader) Read32Bytes() (data [32]byte, err error) {
    _, err = io.ReadFull(b.src, data[:])
    return
}

func (b *BufReader) ReadUint8() (data uint8, err error) {
    d, err := b.ReadBytes(1)
    if err != nil {
        return 0, err
    }
    return d[0], err
}

func (b *BufReader) ReadUint16() (data uint16, err error) {
    d, err := b.ReadBytes(2)
    if err != nil {
        return 0, err
    }
    return binary.LittleEndian.Uint16(d), nil
}

func (b *BufReader) ReadUint32() (data uint32, err error) {
    d, err := b.ReadBytes(4)
    if err != nil {
        return 0, err
    }
    return binary.LittleEndian.Uint32(d), nil
}

func (b *BufReader) ReadUint64() (data uint64, err error) {
    d, err := b.ReadBytes(8)
    if err != nil {
        return 0, err
    }
    return binary.LittleEndian.Uint64(d), nil
}

func (b *BufReader) ReadInt32() (data int32, err error) {
    err = binary.Read(b.src, binary.LittleEndian, &data)
    return 
}

func (b *BufReader) ReadInt64() (data int64, err error) {
    err = binary.Read(b.src, binary.LittleEndian, &data)
    return 
}

var CompactUintError = errors.New("CompactUintError")
// https://btcinformation.org/en/developer-reference#compactsize-unsigned-integers
func (b *BufReader) ReadCompactUint() (data uint64, err error) {
    d, err := b.ReadUint8()
    if err != nil {
        return 0, err
    }
    switch d {
    case 253:
        var d2 uint16
        d2, err = b.ReadUint16()
        if err != nil {
            return 0, err
        }
        if !(d2 >= 253) {
            return 0, CompactUintError
        }
        data = uint64(d2)
    case 254:
        var d2 uint32
        d2, err = b.ReadUint32()
        if err != nil {
            return 0, err
        }
        if !(d2 >= 0x10000) {
            return 0, CompactUintError
        }
        data = uint64(d2)
    case 255:
        data, err = b.ReadUint64()
        if err != nil {
            return 0, err
        }
        if !(data >= 0x100000000) {
            return 0, CompactUintError
        }
    default:
        data, err = uint64(d), nil
    }
    return data, err
}
