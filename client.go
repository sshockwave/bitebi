package main

import (
	"bufio"
	"fmt"
	"os"
)

type CmdApp struct {
	isTerminal bool
	scanner *bufio.Scanner
	blockchain BlockChain
	peer Peer
	name string
}

func NewCmdApp() (app CmdApp) {
	o, _ := os.Stdout.Stat()
	app.isTerminal = (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	app.scanner = bufio.NewScanner(os.Stdin)
	// TODO name: random name
	// TODO init peer
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
		case "stopmining":
			// stop all mining processes
		case "peer": // sk
			// add an address of a peer
		case "createtx":
			// input extra
		case "showbalance":
			// display the balance of an account
		}
		for c.scanner.Scan() {
		}
	}
	if c.scanner.Err() != nil {
		fmt.Println("[ERROR] During scanning, an error occurred: " + c.scanner.Err().Error())
	}
}


func main() {
	app := NewCmdApp()
	app.Serve()
}
