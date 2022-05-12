package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type NetworkIPAddress struct {
	time     uint32
	services uint64
	Ipv6     [16]byte
	Port     uint16
}

func (a *NetworkIPAddress) LoadBuffer(reader utils.BufReader) (err error) {
	a.time, err = reader.ReadUint32()
	if err != nil {
		return
	}
	a.services, err = reader.ReadUint64()
	if err != nil {
		return
	}
	a.Ipv6, err = reader.Read16Bytes()
	if err != nil {
		return
	}
	a.Port, err = reader.ReadUint16()
	return
}

func (a *NetworkIPAddress) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteUint32(a.time)
	if err != nil {
		return
	}
	err = writer.WriteUint64(a.services)
	if err != nil {
		return
	}
	err = writer.WriteBytes(a.Ipv6[:])
	if err != nil {
		return
	}
	err = writer.WriteUint16(a.Port)
	return
}

type AddrMsg struct {
	Addrs []NetworkIPAddress
}

func (a *AddrMsg) LoadBuffer(reader utils.BufReader) (err error) {
	var cnt uint64
	cnt, err = reader.ReadCompactUint()
	if err != nil {
		return
	}
	a.Addrs = make([]NetworkIPAddress, cnt)
	for i := range a.Addrs {
		err = a.Addrs[i].LoadBuffer(reader)
		if err != nil {
			return
		}
	}
	return
}

func (a *AddrMsg) PutBuffer(writer utils.BufWriter) (err error) {
	err = writer.WriteCompactUint(uint64(len(a.Addrs)))
	if err != nil {
		return
	}
	for i := range a.Addrs {
		err = a.Addrs[i].PutBuffer(writer)
		if err != nil {
			return
		}
	}
	return
}
