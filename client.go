package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/sshockwave/bitebi/p2p"
	"github.com/sshockwave/bitebi/utils"
)

type CmdApp struct {
	isTerminal bool
	scanner    *bufio.Scanner
	blockchain BlockChain
	peer       Peer
	hasPeer    bool
	name       string
}

func NewCmdApp() (app CmdApp) {
	o, _ := os.Stdout.Stat()
	app.isTerminal = (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	app.scanner = bufio.NewScanner(os.Stdin)
	app.name = utils.RandomName()
	return
}

func (c *CmdApp) Serve() {
	for {
		if c.isTerminal {
			fmt.Print(">>")
		}
		if !c.scanner.Scan() {
			break
		}
		switch c.scanner.Text() {
		case "mine":
			// create a goroutine that mines
			go c.blockchain.mine(0, 1, c.peer)
		case "stopmining":
			// stop all mining processes
			c.blockchain.Mining = false
		case "peer": // sk
			// add an address of a peer
			if !c.scanner.Scan() {
				break
			}
			addr := c.scanner.Text()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Println("[ERROR] Dialing address " + addr + " failed")
			} else {
				var new_c PeerConnection
				new_c.Conn = conn
				new_c.peer = &c.peer
				go new_c.Serve()
			}
		case "createtx":
			// input extra
		case "showbalance":
			// display the balance of an account
		case "serve":
			if c.hasPeer {
				log.Println("[ERROR] A server is already running!")
			} else {
				var err error
				c.peer, err = NewPeer(&c.blockchain, p2p.GetMainnet(), "0.0.0.0", -1)
				if err != nil {
					fmt.Println("[ERROR] " + err.Error())
				} else {
					c.hasPeer = true
				}
			}
		case "sleep":
			if !c.scanner.Scan() {
				break
			}
			t, err := strconv.Atoi(c.scanner.Text())
			if err != nil {
				log.Println("[ERROR] Time parsing error: " + err.Error())
			}
			time.Sleep(time.Duration(t) * time.Second)
		}
	}
	if c.scanner.Err() != nil {
		log.Println("[ERROR] During scanning, an error occurred: " + c.scanner.Err().Error())
	}
}

func main() {
	utils.RandomInit()
	app := NewCmdApp()
	app.Serve()
}
