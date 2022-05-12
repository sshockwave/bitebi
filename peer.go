package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/sshockwave/bitebi/p2p"
	"github.com/sshockwave/bitebi/utils"
)

type void struct{}
var void_null void

type Peer struct {
	Chain *BlockChain
	Config p2p.NetConfig
	ln net.Listener
	conns map[*PeerConnection]void
	lock sync.RWMutex
}

func (c *PeerConnection) Serve() {
	c.peer.lock.Lock()
	c.peer.conns[c] = void_null
	c.peer.lock.Unlock()
	for {
		err := c.readMessage()
		if err != nil {
			log.Println("[ERROR] " + err.Error())
			break
		}
	}
	c.peer.lock.Lock()
	delete(c.peer.conns, c)
	c.peer.lock.Unlock()
	c.Conn.Close()
}

func (p *Peer) messageLoop() {
	for {
		conn, err := p.ln.Accept()
		if err != nil {
			log.Println("[ERROR] Accepting TCP connections: " + err.Error())
			// handle error
			break
		}
		c := new(PeerConnection)
		c.Conn = conn
		c.peer = p
		go c.Serve()
	}
}

func NewPeer(chain *BlockChain, cfg p2p.NetConfig, host string, port int) (p Peer, err error) {
	p.Chain = chain
	p.Config = cfg
	if port < 0 {
		port = cfg.DefaultPort
	}
	p.ln, err = net.Listen("tcp", host + ":" + strconv.Itoa(port))
	if err != nil {
		return
	}
	log.Println("[INFO] Server listening on " + p.ln.Addr().String())
	go p.messageLoop()
	return
}

func (p *Peer) BroadcastTransaction() {
}

func (p *Peer) BroadcastBlock() {
}

type PeerConnection struct {
	Conn net.Conn
	peer *Peer
}

func (c *PeerConnection) readMessage() (err error) {
	header := make([]byte, 4 + 12 + 4 + 4)
	_, err = io.ReadFull(c.Conn, header)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Println("[ERROR] Error returned when reading header")
		return
	}
	if bytes.Compare(c.peer.Config.StartString[:], header[0:4]) != 0 {
		log.Println("[ERROR] Invalid start string, discarding")
		err = errors.New("invalidStartString")
		return
	}
	command := string(bytes.TrimRight(header[4:16], "\x00"))
	payload_size := binary.LittleEndian.Uint32(header[16:20])
	if payload_size > c.peer.Config.MaxNBits {
		log.Println("[ERROR] Payload too large")
		err = errors.New("payloadTooLarge")
		return
	}
	payload := make([]byte, payload_size)
	_, err = io.ReadFull(c.Conn, payload)
	if err != nil {
		log.Println("[ERROR] Error occurred while reading payload")
		return
	}
	recv_chksum := utils.Sha256Twice(payload)
	if bytes.Compare(recv_chksum[:4], header[20:24]) != 0 {
		log.Println("[ERROR] Checksum do not match in message, discarding")
		return
	}
	reader := utils.NewBufReader(bytes.NewBuffer(payload))
	switch command {
	// Data messages
	// https://developer.bitcoin.org/reference/p2p_networking.html#id1
	case "getheaders":
		// Not yet in plan
	case "headers":
		// Not yet in plan
	case "getblocks":
		// return "inv", at most 500
	case "mempool":
	case "inv":
	case "getdata":
	case "tx":
	case "block":
		// SerializedBlock
	case "merkleblock":
	case "notfound":
		// Not yet in plan

	// Control messages
	case "version":
	case "verack":
	case "ping":
		// Not yet in plan
	case "pong":
		// Not yet in plan
	case "getaddr":
	case "addr":
	case "addrv2":
		// Not yet in plan
	case "filterload":
		// Not yet in plan
	case "filteradd":
		// Not yet in plan
	case "filterclear":
		// Not yet in plan
	case "sendaddrv2":
		// Not yet in plan
	case "sendheaders":
		// Not yet in plan
	case "reject":
		// Not yet in plan
	}
	return
}
