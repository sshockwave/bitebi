package main

import (
	"bufio"
	"fmt"
	"os"

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
			go c.blockchain.mine(0, c.blockchain.Mempool, 1, c.peer)
		case "stopmining":
			// stop all mining processes
			c.blockchain.Mining = false
		case "peer": // sk
			// add an address of a peer
		case "createtx":
			// input extra
		case "showbalance":
			// display the balance of an account
		case "serve":
			if c.hasPeer {
				fmt.Println("A server is already running!")
			} else {
				var err error
				c.peer, err = NewPeer(&c.blockchain, p2p.GetMainnet(), "0.0.0.0", -1)
				if err != nil {
					fmt.Println("[ERROR] " + err.Error())
				} else {
					c.hasPeer = true
				}
			}
		}
	}
	if c.scanner.Err() != nil {
		fmt.Println("[ERROR] During scanning, an error occurred: " + c.scanner.Err().Error())
	}
}

func main() {
	utils.RandomInit()
	app := NewCmdApp()
	app.Serve()
}
