package main

import (
	"bufio"
	"fmt"
	"github.com/sshockwave/bitebi/message"
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
	peer       *Peer
	hasPeer    bool
	name       string
}

func NewCmdApp() (app CmdApp) {
	o, _ := os.Stdout.Stat()
	name := utils.RandomName()
	app.isTerminal = (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	app.scanner = bufio.NewScanner(os.Stdin)
	app.name = name
	app.blockchain = BlockChain{
		Mining:     false,
		ClientName: name,
	}
	log.Printf("[INFO] App initialized with name: " + app.name)
	return
}

func (c *CmdApp) Serve() {
	for {
		if c.isTerminal {
			fmt.Print(">> ")
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
				new_c.peer = c.peer
				go new_c.Serve()
			}
		case "transferAccount":
			// input extra
			var accountName string
			var amount int64

			totalPayment := int64(0)
			tx_In := []message.TxIn{}

			for outPoint, _ := range c.blockchain.UTXO {
				hash := outPoint.Hash
				index := outPoint.Index
				transaction := c.blockchain.TX[hash]
				txOut := transaction.Tx_out[index]
				if string(txOut.Pk_script) == c.name {
					value := txOut.Value
					totalPayment += value
					txIn := message.TxIn{
						Previous_output:  outPoint,
						Signature_script: []byte(c.name),
					}
					tx_In = append(tx_In, txIn)
				}
				if totalPayment >= amount {
					break
				}
			}

			if totalPayment >= amount {
				transaction := message.Transaction{
					Version: 0,
					Tx_in:   tx_In,
					Tx_out: []message.TxOut{
						{Value: amount,
							Pk_script: []byte(accountName)},
						{Value: totalPayment - amount,
							Pk_script: []byte(c.name)},
					},
					Lock_time: 0,
				}
				c.peer.BroadcastTransaction(transaction)
			} else {
				fmt.Println("Warning: No transfer was made, because your don't have enough money.")
			}

		case "showbalance":
			// display the balance of an account
			wallet := int64(0)
			for key, _ := range c.blockchain.UTXO {
				hash := key.Hash
				index := key.Index
				transaction := c.blockchain.TX[hash]
				txOut := transaction.Tx_out[index]
				if string(txOut.Pk_script) == c.name {
					wallet += txOut.Value
				}
			}
			fmt.Println("This client has", wallet, "money")
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
		case "exit":
			break
		case "quit":
			break
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
