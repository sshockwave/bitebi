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

	"github.com/sshockwave/bitebi/message"
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
		command, payload, err := c.readMessage()
		if err != nil {
			log.Println("[ERROR] " + err.Error())
			break
		}
		err = c.dispatchMessage(command, payload)
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

func (p *Peer) BroadcastTransaction(tx message.Transaction) (err error) {
	var b []byte
	b, err = utils.GetBytes(&tx)
	if err != nil {
		log.Printf("[ERROR] Serializing tx: " + err.Error())
		return
	}
	for c := range p.conns {
		c.sendMessage("tx", b)
	}
	return
}

func (p *Peer) BroadcastBlock(blk message.SerializedBlock) (err error) {
	var b []byte
	b, err = utils.GetBytes(&blk)
	if err != nil {
		log.Printf("[ERROR] Serializing tx: " + err.Error())
		return
	}
	for c := range p.conns {
		c.sendMessage("tx", b)
	}
	return
}

type PeerConnection struct {
	Conn net.Conn
	peer *Peer
}

func (c *PeerConnection) readMessage() (command string, payload []byte, err error) {
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
	command = string(bytes.TrimRight(header[4:16], "\x00"))
	payload_size := binary.LittleEndian.Uint32(header[16:20])
	if payload_size > c.peer.Config.MaxNBits {
		log.Println("[ERROR] Payload too large")
		err = errors.New("payloadTooLarge")
		return
	}
	payload = make([]byte, payload_size)
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
	return
}

func (c *PeerConnection) dispatchMessage(command string, payload []byte) (err error) {
	switch command {
	// Data messages
	// https://developer.bitcoin.org/reference/p2p_networking.html#id1
	case "getheaders":
		// Not yet in plan
	case "headers":
		// Not yet in plan
	case "getblocks":
		c.onGetBlocks(payload)
		// return "inv", at most 500
	case "mempool":
		c.onMempool(payload)
	case "inv":
		c.onInv(payload)
	case "getdata":
	case "tx":
		c.onTx(payload)
	case "block":
		// SerializedBlock
	case "merkleblock":
		// Not yet in plan
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

func (c *PeerConnection) sendMessage(command string, payload []byte) (err error) {
	header := utils.NewBufWriter()
	header.WriteBytes(c.peer.Config.StartString[:])
	t := []byte(command)
	header.WriteBytes(t)
	for i := 0; i < 12 - len(t); i++ {
		header.WriteUint8(0)
	}
	header.WriteUint32(uint32(len(payload)))
	chksum := utils.Sha256Twice(payload)
	header.WriteBytes(chksum[:4])
	_, err = c.Conn.Write(header.Collect())
	if err != nil {
		return
	}
	_, err = c.Conn.Write(payload)
	return
}

func (c *PeerConnection) onMempool(data []byte) (err error) {
	inv := make([][]message.Inventory, 0)
	c.peer.Chain.Mtx.Lock()
	cur_pt := make([]message.Inventory, 0)
	inv = append(inv, cur_pt)
	for k := range c.peer.Chain.Mempool {
		if len(cur_pt) == message.InvMaxItemCount {
			cur_pt := make([]message.Inventory, 0)
			inv = append(inv, cur_pt)
		}
		cur_pt = append(cur_pt, message.Inventory{message.MSG_TX, k})
	}
	c.peer.Chain.Mtx.Unlock()
	for _, v := range inv {
		msg := message.InvMsg{v}
		data, _ = utils.GetBytes(&msg)
		err = c.sendMessage("inv", data)
		if err != nil {
			return
		}
	}
	return
}

func (c *PeerConnection) onGetBlocks(data []byte) (err error) {
	reader := bytes.NewBuffer(data)
	msg, err := message.NewGetBlocksMsg(reader)
	if err != nil {
		return
	}
	// ignoring msg.Version
	var commonHeight int
	c.peer.Chain.Mtx.Lock()
	for _, hash := range msg.BlockHeaderHashes {
		if h, ok := c.peer.Chain.Height[hash]; ok {
			commonHeight = h
			c.peer.Chain.Mtx.Unlock()
			break
		}
	}
	c.peer.Chain.Mtx.Unlock()
	inv := make([]message.Inventory, 0)
	cnt := 0
	for i := commonHeight + 1; i < len(c.peer.Chain.Block); i++ {
		if c.peer.Chain.Block[i].HeaderHash == msg.StopHash {
			break
		}
		inv = append(inv, message.Inventory{message.MSG_BLOCK, c.peer.Chain.Block[i].HeaderHash})
		cnt += 1
		if cnt == message.InvMaxItemCount {
			break
		}
	}
	invmsg := message.InvMsg{inv}
	invbytes, err := utils.GetBytes(&invmsg)
	if err != nil {
		return
	}
	err = c.sendMessage("inv", invbytes)
	return
}

func (c *PeerConnection) onInv(data []byte) (err error) {
	reader := utils.NewBufReader(bytes.NewBuffer(data))
	invmsg, err := message.NewInvMsg(reader)
	if err != nil {
		return
	}
	retmsg := message.InvMsg{make([]message.Inventory, 0)}
	c.peer.lock.RLock()
	c.peer.Chain.Mtx.Lock()
	for _, v := range invmsg.Inv {
		switch v.Type {
		case message.MSG_BLOCK:
			ok := false
			if !ok {
				_, ok = c.peer.Chain.Height[v.Hash]
			}
			if !ok {
				_, ok = c.peer.history[v.Hash]
			}
			if !ok {
				retmsg.Inv = append(retmsg.Inv, v)
			}
		case message.MSG_TX:
			ok := false
			_, ok = c.peer.Chain.TX[v.Hash]
			if !ok {
				retmsg.Inv = append(retmsg.Inv, v)
			}
		default:
			log.Printf("[ERROR] Unknown inv type: " + strconv.Itoa(int(v.Type)))
		}
	}
	c.peer.Chain.Mtx.Unlock()
	c.peer.lock.RUnlock()
	if len(retmsg.Inv) > 0 {
		var data []byte
		data, err = utils.GetBytes(&retmsg)
		if err != nil {
			return
		}
		err = c.sendMessage("getdata", data)
	}
	return
}

func (c *PeerConnection) onTx(data []byte) (err error) {
	reader := utils.NewBufReader(bytes.NewBuffer(data))
	var tx message.Transaction
	err = tx.LoadBuffer(reader)
	if err != nil {
		return
	}
	var flag bool
	var hash [32]byte
	hash, err = utils.GetHash(&tx)
	if err != nil {
		return err
	}
	c.peer.Chain.Mtx.Lock()
	_, flag = c.peer.Chain.TX[hash]
	c.peer.Chain.Mtx.Unlock()
	if !flag {
		c.peer.Chain.addTransaction(tx)
		err = c.peer.BroadcastTransaction(tx)
	}
	return
}
