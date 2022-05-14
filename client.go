package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sshockwave/bitebi/message"

	"github.com/sshockwave/bitebi/p2p"
	"github.com/sshockwave/bitebi/utils"
)

type CmdApp struct {
	isTerminal bool
	scanner    *bufio.Scanner
	TokenScanner *bufio.Scanner
	blockchain BlockChain
	peer       *Peer
	hasPeer    bool
	name string
}

func NewCmdApp() (app CmdApp) {
	o, _ := os.Stdout.Stat()
	app.isTerminal = (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	app.scanner = bufio.NewScanner(os.Stdin)
	app.blockchain.init()
	app.name = utils.RandomName()
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
		c.TokenScanner = bufio.NewScanner(strings.NewReader(c.scanner.Text()))
		c.TokenScanner.Split(bufio.ScanWords)
		if !c.TokenScanner.Scan() {
			fmt.Println("Empty command.")
			continue
		}
		switch c.TokenScanner.Text() {
		case "mine":
			// create a goroutine that mines
			go c.blockchain.mine(0, 0x03001000, c.peer, []byte(c.name))
		case "stopmining":
			// stop all mining processes
			c.blockchain.PauseMining()
		case "resumemining":
			// stop all mining processes
			c.blockchain.ResumeMining()
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
		case "transfer":
			// input extra
			var fromAccount string
			var accountName string
			var amount int64
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: transfer <from> <to> <amount>")
				continue
			}
			fromAccount = c.TokenScanner.Text()
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: transfer <from> <to> <amount>")
				continue
			}
			accountName = c.TokenScanner.Text()
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: transfer <from> <to> <amount>")
				continue
			}
			tmp, err := strconv.Atoi(c.TokenScanner.Text())
			if err != nil {
				log.Println("[ERROR] Input amount is not an integer")
				continue
			}
			amount = int64(tmp)

			totalPayment := int64(0)
			tx_In := []message.TxIn{}

			for outPoint, _ := range c.blockchain.UTXO {
				hash := outPoint.Hash
				index := outPoint.Index
				transaction := c.blockchain.TX[hash]
				txOut := transaction.Tx_out[index]
				if string(txOut.Pk_script) == fromAccount {
					value := txOut.Value
					totalPayment += value
					txIn := message.TxIn{
						Previous_output:  outPoint,
						Signature_script: []byte(fromAccount),
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
							Pk_script: []byte(fromAccount)},
					},
					Lock_time: 0,
				}
				c.peer.BroadcastTransaction(transaction)
			} else {
				log.Println("[ERROR] No transfer was made, because your don't have enough money.")
			}

		case "showbalance":
			chosen_account := c.name
			if c.TokenScanner.Scan() {
				chosen_account = c.TokenScanner.Text()
			}
			// display the balance of an account
			wallet := int64(0)
			c.blockchain.Mtx.Lock()
			for key, val := range c.blockchain.UTXO {
				if val == false {
					continue
				}
				hash := key.Hash
				index := key.Index
				transaction := c.blockchain.TX[hash]
				txOut := transaction.Tx_out[index]
				if string(txOut.Pk_script) == chosen_account {
					wallet += txOut.Value
				}
			}
			c.blockchain.Mtx.Unlock()
			fmt.Println("Client ", chosen_account, " has ", wallet, " money")
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
		default:
			log.Println("Unknown command: \"" + c.TokenScanner.Text() + "\"")
		case "showpeer":
			for _, addr := range c.peer.GetPeerList() {
				fmt.Println(addr)
			}
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
