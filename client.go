package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/p2p"
	"github.com/sshockwave/bitebi/utils"
)

type CmdApp struct {
	isTerminal bool
	LineScanner    *bufio.Scanner
	TokenScanner *bufio.Scanner
	blockchain BlockChain
	peer       *Peer
	hasPeer    bool
	name string
}

func NewCmdApp() (app CmdApp) {
	o, _ := os.Stdout.Stat()
	var inputfile string
	flag.StringVar(&inputfile, "input", "-", "Input File")
	flag.Parse()
	app.isTerminal = (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	if inputfile == "-" {
		app.LineScanner = bufio.NewScanner(os.Stdin)
	} else {
		app.isTerminal = false
		f, err := os.Open(inputfile)
		if err != nil {
			log.Fatalf("[ERROR] cannot open " + inputfile + ", " + err.Error())
		}
		app.LineScanner = bufio.NewScanner(f)
	}
	app.blockchain.init()
	app.name = utils.RandomName()
	log.Printf("[INFO] App initialized with name: " + app.name)
	return
}

func (c *CmdApp) Serve() {
	fmt.Println("Welcome!")
	fmt.Println("To get started, start a peer by 'serve <port>'")
	fmt.Println("Then add some peer with 'peer <addr>'")
	for {
		if c.isTerminal {
			fmt.Print(">> ")
		}
		if !c.LineScanner.Scan() {
			if c.LineScanner.Err() != nil {
				log.Println("[ERROR] During scanning, an error occurred: " + c.LineScanner.Err().Error())
			} else if !c.isTerminal {
				log.Println("[INFO] Input completed. Entering infinite loop.")
				var wg sync.WaitGroup
				wg.Add(1)
				wg.Wait()
			}
			break
		}
		c.TokenScanner = bufio.NewScanner(strings.NewReader(c.LineScanner.Text()))
		c.TokenScanner.Split(bufio.ScanWords)
		if !c.TokenScanner.Scan() {
			log.Println("[INFO] Empty command.")
			continue
		}
		if !c.hasPeer && c.TokenScanner.Text() != "serve" {
			log.Println("[ERROR] A peer has not been initiated.")
			continue
		}
		switch c.TokenScanner.Text() {
		case "mine":
			// create a goroutine that mines
			go c.blockchain.mine(0, 0x1E100000, c.peer, []byte(c.name))
		case "stopmining":
			// stop all mining processes
			c.blockchain.PauseMining()
		case "resumemining":
			// stop all mining processes
			c.blockchain.ResumeMining()
		case "peer": // sk
			// add an address of a peer
			if !c.TokenScanner.Scan() {
				break
			}
			addr := c.TokenScanner.Text()
			conn, err := c.peer.Dial(addr)
			if err != nil {
				log.Println("[ERROR] Dialing address " + addr + " failed")
			} else {
				c.peer.NewConn(conn)
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
				c.blockchain.addTransaction(transaction)
				c.blockchain.refreshMining()
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
				nc := p2p.GetMainnet()
				if c.TokenScanner.Scan() {
					var err error
					nc.DefaultPort, err = strconv.Atoi(c.TokenScanner.Text())
					if err != nil {
						log.Println("[ERROR] the port number should be an integer")
						continue
					}
				}
				var err error
				c.peer, err = NewPeer(&c.blockchain, nc, "127.0.0.1", -1)
				if err != nil {
					fmt.Println("[ERROR] " + err.Error())
				} else {
					c.hasPeer = true
				}
			}
		case "sleep":
			if !c.TokenScanner.Scan() {
				break
			}
			t, err := strconv.Atoi(c.TokenScanner.Text())
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
}

func main() {
	utils.RandomInit()
	app := NewCmdApp()
	app.Serve()
}
