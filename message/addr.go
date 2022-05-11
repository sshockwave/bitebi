package message

import (
	"github.com/sshockwave/bitebi/utils"
)

type NetworkIPAddress struct {
	time     uint32
	services uint64
	ipv6     [16]byte
}

func NewNetworkIPAddress(reader utils.BufReader) {
}
