package main

import (
	"bufio"
	"crypto/dsa"
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
	isTerminal   bool
	LineScanner  *bufio.Scanner
	TokenScanner *bufio.Scanner
	blockchain   BlockChain
	peer         *Peer
	hasPeer      bool
	Wallet       Wallet
}

func NewCmdApp() (app CmdApp) {
	o, _ := os.Stdin.Stat()
	var inputfile string
	flag.StringVar(&inputfile, "input", "-", "Input File")
	flag.Parse()
	app.isTerminal = (o.Mode() & os.ModeCharDevice) != 0
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
	app.Wallet.Init(&app.blockchain)
	app.blockchain.init(&app.Wallet)
	privateKey := GenPrivKey()
	app.Wallet.AddPrivKey("self", privateKey)
	log.Printf("[INFO] PrivKey: " + string(SK2Bytes(privateKey)))
	log.Printf("[INFO] PubKey: " + string(PK2Bytes(privateKey.PublicKey)))
	return
}

func (c *CmdApp) Serve() {
	if c.isTerminal {
		fmt.Println("Welcome!")
		fmt.Println("To get started, start a peer by 'serve <port>'")
		fmt.Println("Then add some peer with 'peer <addr>'")
	}
	for {
		if c.isTerminal {
			fmt.Print(">> ")
		}
		if !c.LineScanner.Scan() {
			if c.LineScanner.Err() != nil {
				log.Println("[ERROR] During scanning, an error occurred:", c.LineScanner.Err().Error())
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
			// Examples: easiest(0x20ffffff), hardest(0x03000000)
			pkscript := GenerateP2PKHPkScript(c.Wallet.Pubkey["self"])
			go c.blockchain.mine(0, c.peer.Config.MaxNBits, c.peer, pkscript)
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
				log.Printf("[ERROR] Dialing address %v failed: %v", addr, err)
			} else {
				c.peer.NewConn(conn)
			}
		case "addpk":
			var name string
			var pkstring string
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: addpk <name> <public key>")
				continue
			}
			name = c.TokenScanner.Text()
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: addpk <name> <public key>")
				continue
			}
			pkstring = c.TokenScanner.Text()
			pk := Bytes2PK([]byte(pkstring))
			c.Wallet.AddPubKey(name, pk)
		case "addsk":
			var name string
			var skstring string
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: addpk <name> <public key>")
				continue
			}
			name = c.TokenScanner.Text()
			if !c.TokenScanner.Scan() {
				log.Println("[ERROR] Usage: addpk <name> <public key>")
				continue
			}
			skstring = c.TokenScanner.Text()
			sk := Bytes2SK([]byte(skstring))
			c.Wallet.AddPrivKey(name, sk)
		case "transfer":
			// input extra
			var fromAccount string
			var fromAccount_SK dsa.PrivateKey
			var accountName string
			var accountName_PK dsa.PublicKey
			var amount int64
			var ok bool
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

			accountName_PK, ok = c.Wallet.Pubkey[accountName]
			if !ok {
				log.Printf("[ERROR] No known pubkey for %v", accountName)
				continue
			}
			fromAccount_SK, ok = c.Wallet.GetSK(fromAccount) // TODO: this line is not thread safe
			if !ok {
				log.Printf("[ERROR] No known privkey for %v", fromAccount)
				continue
			}
			totalPayment, outpoints := c.Wallet.MakeTxIn(fromAccount, amount)
			if totalPayment < amount {
				log.Println("[ERROR] No transfer was made, because your don't have enough money.")
				continue
			}
			tx_In := []message.TxIn{}
			for _, o := range outpoints {
				tx_In = append(tx_In, message.TxIn{ Previous_output: o })
			}
			oput := []message.TxOut{{Value: amount, Pk_script: GenerateP2PKHPkScript(accountName_PK)}}
			if totalPayment > amount {
				oput = append(oput, message.TxOut{
					Value:                             totalPayment - amount,
					Pk_script: GenerateP2PKHPkScript(fromAccount_SK.PublicKey),
				})
			}
			if totalPayment >= amount {
				transaction := message.Transaction{
					Version:   0,
					Tx_in:     tx_In,
					Tx_out:    oput,
					Lock_time: 0,
				}

				signature := SignTransaction(fromAccount_SK, transaction)
				for i := 0; i < len(transaction.Tx_in); i++ {
					transaction.Tx_in[i].Signature_script = signature
				}

				c.blockchain.Mtx.Lock()
				c.blockchain.addTransaction(transaction)
				mempool_size := len(c.blockchain.Mempool)
				c.blockchain.Mtx.Unlock()
				c.Wallet.RemoveUTXO(fromAccount, outpoints)
				if mempool_size < 100 {
					c.blockchain.refreshMining()
				}
				c.peer.BroadcastTransaction(transaction)
			}

		case "showbalance":
			chosen_account := "self"
			if c.TokenScanner.Scan() {
				chosen_account = c.TokenScanner.Text()
			}
			// display the balance of an account
			val := c.Wallet.GetBalance(chosen_account)
			log.Printf("Client %v has %v satoshis", chosen_account, val)
		case "serve":
			if c.hasPeer {
				log.Println("[ERROR] A server is already running!")
			} else {
				nc := p2p.GetBitebinet()
				if c.TokenScanner.Scan() {
					var err error
					nc.DefaultPort, err = strconv.Atoi(c.TokenScanner.Text())
					if err != nil {
						log.Println("[ERROR] the port number should be an integer")
						continue
					}
				}
				var err error
				c.peer, err = NewPeer(&c.blockchain, nc, "0.0.0.0", -1)
				if err != nil {
					fmt.Println("[ERROR]", err)
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
				log.Println("[ERROR] Time parsing error:", err.Error())
			}
			time.Sleep(time.Duration(t) * time.Second)
		default:
			log.Printf("Unknown command: \"%v\"", c.TokenScanner.Text())
		case "showpeer":
			for _, addr := range c.peer.GetPeerList() {
				fmt.Println(addr)
			}
		case "stat":
			last_cnt := 0
			time_int := 200 // ms
			tot_cnt := 0
			loop_cnt := 0
			for {
				c.peer.lock.Lock()
				peer_cnt := len(c.peer.conns) + 1
				c.peer.lock.Unlock()
				c.blockchain.Mtx.Lock()
				block_cnt := len(c.blockchain.Block)
				unconfirmed_tx_cnt := len(c.blockchain.Mempool)
				confirmed_tx_cnt := len(c.blockchain.TX) - unconfirmed_tx_cnt - len(c.blockchain.Block) + 1
				if confirmed_tx_cnt > last_cnt {
					tot_cnt += confirmed_tx_cnt - last_cnt
				}
				last_cnt = confirmed_tx_cnt
				loop_cnt += 1
				c.blockchain.Mtx.Unlock()
				fmt.Printf(
					"Stats: %v nodes; %v blocks; %v tx; %v unconfirmed tx; %v valid tx / sec\r",
					peer_cnt,
					block_cnt,
					confirmed_tx_cnt,
					unconfirmed_tx_cnt,
					float64(tot_cnt)/float64(loop_cnt)*1000/float64(time_int),
				)
				time.Sleep(time.Duration(time_int) * time.Millisecond)
			}
		}
	}
}

func main() {
	utils.RandomInit()
	app := NewCmdApp()
	app.Serve()
}
