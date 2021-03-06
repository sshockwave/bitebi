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

	reuse "github.com/libp2p/go-reuseport"
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
	orphans Orphans
}

func ConnectionToAddr(c net.Addr) net.TCPAddr {
	addr, ok := c.(*net.TCPAddr)
	if !ok {
		log.Fatalf("[ERROR] Listener should have been tcp")
	}
	return *addr
}

func (c *PeerConnection) Serve() {
	c.peer.lock.Lock()
	c.peer.conns[c] = void_null
	c.peer.lock.Unlock()
	// TODO: version message
	c.sendMessage("getaddr", []byte{})
	c.doBlockSync()
	for {
		command, payload, err := c.readMessage()
		if err == io.EOF {
			log.Printf("[INFO] Connection to %v ended.", c.Conn.RemoteAddr())
			break
		} else if err != nil {
			log.Println("[ERROR]", err)
			break
		}
		err = c.dispatchMessage(command, payload)
		if err != nil {
			log.Println("[ERROR]", err)
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
			log.Println("[ERROR] Accepting TCP connections:", err)
			break
		}
		p.NewConn(conn)
	}
}

func NewPeer(chain *BlockChain, cfg p2p.NetConfig, host string, port int) (p *Peer, err error) {
	p = new(Peer)
	p.Chain = chain
	p.Config = cfg
	p.conns = make(map[*PeerConnection]void)
	p.orphans.Init(chain)
	if port < 0 {
		port = cfg.DefaultPort
	}
	p.ln, err = reuse.Listen("tcp", host + ":" + strconv.Itoa(port))
	if err != nil {
		return
	}
	log.Println("[INFO] Server listening on", p.ln.Addr())
	go p.messageLoop()
	return
}

func (p *Peer) Dial(addr string) (net.Conn, error) {
	return reuse.Dial("tcp", p.ln.Addr().String(), addr)
}

func (p *Peer) BroadcastTransaction(tx message.Transaction) (err error) {
	var b []byte
	b, err = utils.GetBytes(&tx)
	if err != nil {
		log.Println("[ERROR] Serializing tx:", err)
		return
	}
	p.lock.RLock()
	for c := range p.conns {
		c.sendMessage("tx", b)
	}
	p.lock.RUnlock()
	return
}

func (p *Peer) BroadcastBlock(blk message.SerializedBlock) (err error) {
	var b []byte
	b, err = utils.GetBytes(&blk)
	if err != nil {
		log.Println("[ERROR] Serializing tx:", err)
		return
	}
	p.lock.RLock()
	for c := range p.conns {
		c.sendMessage("block", b)
	}
	p.lock.RUnlock()
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
		c.onGetData(payload)
	case "tx":
		c.onTx(payload)
	case "block":
		c.onBlock(payload)
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
		c.onGetAddr(payload)
	case "addr":
		c.onAddr(payload)
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
		cur_pt = append(cur_pt, message.Inventory{Type: message.MSG_TX, Hash: k})
	}
	c.peer.Chain.Mtx.Unlock()
	for _, v := range inv {
		msg := message.InvMsg{Inv: v}
		data, _ = utils.GetBytes(&msg)
		err = c.sendMessage("inv", data)
		if err != nil {
			return
		}
	}
	return
}

func (c *PeerConnection) onGetBlocks(data []byte) (err error) {
	var msg message.GetBlocksMsg
	err = msg.LoadBuffer(utils.NewBufReader(bytes.NewBuffer(data)))
	if err != nil {
		return
	}
	// ignoring msg.Version
	var commonHeight int
	c.peer.Chain.Mtx.Lock()
	for _, hash := range msg.BlockHeaderHashes {
		if h, ok := c.peer.Chain.Height[hash]; ok {
			commonHeight = h
			break
		}
	}
	inv := make([]message.Inventory, 0)
	cnt := 0
	for i := commonHeight + 1; i < len(c.peer.Chain.Block); i++ {
		if c.peer.Chain.Block[i].HeaderHash == msg.StopHash {
			break
		}
		inv = append(inv, message.Inventory{Type: message.MSG_BLOCK, Hash: c.peer.Chain.Block[i].HeaderHash})
		cnt += 1
		if cnt == message.InvMaxItemCount {
			break
		}
	}
	c.peer.Chain.Mtx.Unlock()
	invmsg := message.InvMsg{Inv: inv}
	invbytes, err := utils.GetBytes(&invmsg)
	if err != nil {
		return
	}
	err = c.sendMessage("inv", invbytes)
	return
}

func (c *PeerConnection) onInv(data []byte) (err error) {
	reader := utils.NewBufReader(bytes.NewBuffer(data))
	var invmsg message.InvMsg
	err = invmsg.LoadBuffer(reader)
	if err != nil {
		return
	}
	retmsg := message.InvMsg{Inv: make([]message.Inventory, 0)}
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
				if node, tmp := c.peer.orphans.nodes[v.Hash]; tmp {
					ok = node.blk != nil
				}
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
			log.Println("[ERROR] Unknown inv type: ", v.Type)
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
	if !flag {
		c.peer.Chain.addTransaction(tx)
	}
	c.peer.Chain.Mtx.Unlock()
	if !flag {
		err = c.peer.BroadcastTransaction(tx)
	}
	return
}

func (c *PeerConnection) onGetData(data []byte) (err error) {
	reader := utils.NewBufReader(bytes.NewBuffer(data))
	var msg message.InvMsg
	err = msg.LoadBuffer(reader)
	if err != nil {
		return err
	}
	for _, v := range msg.Inv {
		switch v.Type {
		case message.MSG_BLOCK:
			var blk message.SerializedBlock
			c.peer.Chain.Mtx.Lock()
			h, ok := c.peer.Chain.Height[v.Hash]
			if ok {
				blk = c.peer.Chain.Block[h]
			}
			if !ok {
				var node *orphanNode
				node, ok = c.peer.orphans.nodes[v.Hash]
				if ok {
					if node.blk != nil {
						blk = *node.blk
					} else {
						ok = false
					}
				}
			}
			c.peer.Chain.Mtx.Unlock()
			if ok {
				data, _ := utils.GetBytes(&blk)
				c.sendMessage("block", data)
			}
		case message.MSG_TX:
			var tx message.Transaction
			var ok bool
			c.peer.Chain.Mtx.Lock()
			tx, ok = c.peer.Chain.TX[v.Hash]
			c.peer.Chain.Mtx.Unlock()
			if ok {
				data, _ := utils.GetBytes(&tx)
				c.sendMessage("tx", data)
			}
		}
	}
	return
}

var BlockTTL uint64 = 600 // seconds
func (c *PeerConnection) onBlock(data []byte) (err error) {
	var blk message.SerializedBlock
	reader := utils.NewBufReader(bytes.NewBuffer(data))
	err = blk.LoadBuffer(reader)
	if err != nil {
		return
	}
	c.peer.orphans.AddBlock(&blk)
	go c.peer.orphans.RemoveBlock(blk.HeaderHash, BlockTTL)
	chain := c.peer.orphans.GetLongestChain(blk.HeaderHash)
	if len(chain) > 0 {
		c.peer.Chain.Mtx.Lock()
		hei, ok := c.peer.Chain.Height[chain[0].Header.Previous_block_header_hash]
		if ok {
			// not orphaned, check if longer than current chain
			ok = hei + 1 + len(chain) > len(c.peer.Chain.Block)
		}
		c.peer.Chain.Mtx.Unlock()
		if ok {
			chain2 := make([]message.SerializedBlock, len(chain))
			for i, v := range chain {
				chain2[i] = *v
			}
			ok = c.peer.Chain.addBlock(hei + 1, chain2)
		}
		if ok {
			for _, v := range chain {
				c.peer.orphans.RemoveBlock(v.HeaderHash, 0)
			}
		}
	} else {
		err = c.doBlockSync()
	}
	return
}

func (c *PeerConnection) doBlockSync() (err error) {
	var msg message.GetBlocksMsg
	arr := make([][32]byte, 0)
	c.peer.Chain.Mtx.Lock()
	for i, j := len(c.peer.Chain.Block) - 1,  1; i > 0; i, j = i - j, j * 2 {
		arr = append(arr, c.peer.Chain.Block[i].HeaderHash)
	}
	c.peer.Chain.Mtx.Unlock()
	msg.BlockHeaderHashes = arr
	var data []byte
	data, err = utils.GetBytes(&msg)
	if err != nil {
		return
	}
	err = c.sendMessage("getblocks", data)
	return
}

func (p *Peer) GetPeerList() (arr []net.TCPAddr) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	arr = make([]net.TCPAddr, len(p.conns))
	i := 0
	for c := range p.conns {
		arr[i] = ConnectionToAddr(c.Conn.RemoteAddr())
		i++
	}
	return
}

var unexpectedPayload = errors.New("unexpectedPayload")
func (c *PeerConnection) onGetAddr(payload []byte) (err error) {
	if len(payload) > 0 {
		return unexpectedPayload
	}
	peer_list := c.peer.GetPeerList()
	arr := make([]message.NetworkIPAddress, len(peer_list))
	for i, addr := range peer_list {
		copy(arr[i].Ipv6[:], addr.IP.To16())
		arr[i].Port = uint16(addr.Port)
	}
	addrmsg := message.AddrMsg{Addrs: arr}
	var raw_data []byte
	raw_data, err = utils.GetBytes(&addrmsg)
	if err != nil {
		return
	}
	err = c.sendMessage("addr", raw_data)
	return
}

func (p *Peer) NewConn(conn net.Conn) {
	var new_c PeerConnection
	new_c.Conn = conn
	new_c.peer = p
	go new_c.Serve()
	log.Println("[INFO] New connection from", conn.RemoteAddr())
}

func (c *PeerConnection) onAddr(data []byte) (err error) {
	p := c.peer
	reader := utils.NewBufReader(bytes.NewBuffer(data))
	var msg message.AddrMsg
	err = msg.LoadBuffer(reader)
	if err != nil {
		return
	}
	filtered := make([]message.NetworkIPAddress, 0)
	p.lock.RLock()
	peer_list := p.GetPeerList()
	for _, v := range msg.Addrs {
		// check if we are already connected
		found := false
		for _, addr := range peer_list {
			if bytes.Compare(v.Ipv6[:], addr.IP.To16()) == 0 && v.Port == uint16(addr.Port) {
				found = true
				break
			}
		}
		if !found {
			addr := ConnectionToAddr(p.ln.Addr())
			if bytes.Compare(v.Ipv6[:], addr.IP.To16()) == 0 && v.Port == uint16(addr.Port) {
				found = true
			}
		}
		if !found {
			addr := ConnectionToAddr(c.Conn.LocalAddr())
			if bytes.Compare(v.Ipv6[:], addr.IP.To16()) == 0 && v.Port == uint16(addr.Port) {
				found = true
			}
		}
		if !found {
			tcpaddr := net.TCPAddr{IP: v.Ipv6[:], Port: int(v.Port)}
			conn, err := p.Dial(tcpaddr.String())
			if err != nil {
				log.Printf("[WARNING] Connection to %v failed.", tcpaddr)
			} else {
				filtered = append(filtered, v)
				p.NewConn(conn)
			}
		}
	}
	if len(filtered) > 0 {
		addrmsg := message.AddrMsg{Addrs: filtered}
		var raw_data []byte
		raw_data, err = utils.GetBytes(&addrmsg)
		if err != nil {
			return
		}
		for c := range p.conns {
			c.sendMessage("addr", raw_data)
			// err is ignored
		}
	}
	p.lock.RUnlock()
	return
}
