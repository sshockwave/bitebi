package main

import (
	"log"
	"net"
	"strconv"

	"github.com/sshockwave/bitebi/p2p"
)

type Peer struct {
	Chain *BlockChain
	Config p2p.NetConfig
	ln net.Listener
}

func (p *Peer) handleConnection(conn net.Conn) {
	// TODO
}

func (p *Peer) messageLoop() {
	for {
		conn, err := p.ln.Accept()
		if err != nil {
			// handle error
			break
		}
		go p.handleConnection(conn)
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

func (p *Peer) broadcastTransaction() {
}

func (p *Peer) broadcastBlock() {
}

