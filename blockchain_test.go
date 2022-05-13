package main

import (
	"bytes"
	"github.com/sshockwave/bitebi/message"
	"reflect"
	"sync"
	"testing"

	"github.com/sshockwave/bitebi/utils"
)

var blk1 message.Block = message.Block{
	Version:                    -232,
	Previous_block_header_hash: [32]byte{3, 3, 2, 2, 1, 4},
	Merkle_root_hash:           [32]byte{1, 2, 3, 4, 55, 66, 77},
	Time:                       10101010,
	NBits:                      10,
	Nonce:                      0x7f7f7f7f,
}

var blockchain BlockChain = BlockChain{
	Block:      []message.SerializedBlock{},
	Mtx:        sync.Mutex{},
	TX:         map[[32]byte]message.Transaction{},
	Mempool:    map[[32]byte]message.Transaction{},
	Mining:     true,
	Height:     map[[32]byte]int{},
	UTXO:       map[message.Outpoint]bool{},
	ClientName: []byte{12, 45},
}

var tx1 message.Transaction = message.Transaction{
	Version: 1997,
	Tx_in: []message.TxIn{
		{
			Previous_output: message.Outpoint{
				Hash:  [32]byte{33, 22, 0, 11},
				Index: 12,
			},
			Script_bytes:     4,
			Signature_script: []byte{22, 1, 1, 4},
		},
		{
			Previous_output: message.Outpoint{
				Hash:  [32]byte{8, 2, 6, 3},
				Index: 7,
			},
			Script_bytes:     5,
			Signature_script: []byte{1, 2, 3, 4, 5},
		},
	},
	Tx_out: []message.TxOut{
		{
			Value:     22555343,
			Pk_script: []byte{12, 3, 7, 5},
		},
		{
			Value:     -33,
			Pk_script: []byte{0, 0, 3, 0, 4},
		},
		{
			Value:     22,
			Pk_script: []byte{},
		},
	},
	Lock_time: 72,
}

func TestVerifyTransaction(t *testing.T) {
	if blockchain.verifyTransaction(tx1) {
		t.Fatalf("It should return false, but it return true")
	}
}

func TestAddTransaction(t *testing.T) {
	blockchain.addTransaction(tx1)
	hash, _ := utils.GetHash(&tx1)
	tx2 := blockchain.Mempool[hash]
	if utils.GetHash(&tx1) != utils.GetHash(&tx2) {
		t.Fatalf("Failed to add transaction tx1")
	}
}

func TestVerifyBlock(t *testing.T) {

}

func TestAddBlock(t *testing.T) {

}

func doSerializationTest(in utils.BinaryWritable, out utils.BinaryReadable, t *testing.T) {
	b, err := utils.GetBytes(in)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	reader := utils.NewBufReader(bytes.NewBuffer(b))
	err = out.LoadBuffer(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected equal: %v = %v", in, out)
	}
}

func TestTxSerializing(t *testing.T) {
	b, err := utils.GetBytes(&tx1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var tx2 message.Transaction
	reader := utils.NewBufReader(bytes.NewBuffer(b))
	err = tx2.LoadBuffer(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(tx1, tx2) {
		t.Fatalf("Expected equal: %v = %v", tx1, tx2)
	}
}

var blk1 message.Block = message.Block{
	Version:                    -232,
	Previous_block_header_hash: [32]byte{3, 3, 2, 2, 1, 4},
	Merkle_root_hash:           [32]byte{1, 2, 3, 4, 55, 66, 77},
	Time:                       10101010,
	NBits:                      10,
	Nonce:                      0x7f7f7f7f,
}

func TestBlockHeaderSerialization(t *testing.T) {
	var blk2 message.Block
	doSerializationTest(&blk1, &blk2, t)
}
