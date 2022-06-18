package main

import (
	"crypto/dsa"
	"fmt"
	"sync"
	"testing"

	"github.com/sshockwave/bitebi/message"

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

func getHashBlock(b message.Block) [32]byte {
	hash, _ := utils.GetHash(&b)
	return hash
}

var sblk1 message.SerializedBlock = message.SerializedBlock{
	Header:     blk1,
	HeaderHash: getHashBlock(blk1),
	Txns:       []message.Transaction{tx1},
}

var blockchain BlockChain = BlockChain{
	Block:       []message.SerializedBlock{},
	Mtx:         sync.Mutex{},
	TX:          map[[32]byte]message.Transaction{},
	Mempool:     map[[32]byte]message.Transaction{},
	MineVersion: 1,
	Height:      map[[32]byte]int{},
	UTXO:        map[message.Outpoint]bool{},
}

var tx1 message.Transaction = message.Transaction{
	Version: 1997,
	Tx_in: []message.TxIn{
		{
			Previous_output: message.Outpoint{
				Hash:  [32]byte{33, 22, 0, 11},
				Index: 12,
			},
			Signature_script: []byte{22, 1, 1, 4},
		},
		{
			Previous_output: message.Outpoint{
				Hash:  [32]byte{8, 2, 6, 3},
				Index: 7,
			},
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

var pk_script1 []byte = []byte("OP EQUALVERIFY")
var signature_script1 []byte = []byte("199902*199902")

func TestVerifyTxSignature1(t *testing.T) {
	pass := blockchain.verifyScripts(tx1, signature_script1, pk_script1)
	fmt.Println(pass)
	if !pass {
		t.Fatalf("It should pass, but it doesn't.")
	}
}

var sk dsa.PrivateKey = GenPrivKey()
var pk dsa.PublicKey = sk.PublicKey

var pk_script2 []byte = []byte(string(PK2Bytes(pk)) + "*" + "OP CHECKSIG")
var signature_script2 []byte = []byte(SignTransaction(sk, tx1))

func TestVerifyTxSignature2(t *testing.T) {
	success := blockchain.verifyScripts(tx1, signature_script2, pk_script2)

	fmt.Println(success)
	if !success {
		t.Fatalf("It should pass, but it doesn't.")
	}
}

var pk_script3 []byte = []byte("OP DUP" + "*" + "OP EQUAL" + "*" + "OP VERIFY")
var signature_script3 []byte = []byte("777")

func TestVerifyTxSignature3(t *testing.T) {
	success := blockchain.verifyScripts(tx1, signature_script3, pk_script3)

	fmt.Println(success)
	if !success {
		t.Fatalf("It should pass, but it doesn't.")
	}
}

var sk1 dsa.PrivateKey = GenPrivKey()
var pk1 dsa.PublicKey = sk1.PublicKey

var sk2 dsa.PrivateKey = GenPrivKey()
var pk2 dsa.PublicKey = sk2.PublicKey

var sk3 dsa.PrivateKey = GenPrivKey()
var pk3 dsa.PublicKey = sk3.PublicKey

var pk_script4 []byte = []byte("2" + "*" + string(PK2Bytes(pk1)) + "*" + string(PK2Bytes(pk2)) + "*" + string(PK2Bytes(pk3)) + "*" + "3" + "*" + "OP CHECKMULTISIG")
var signature_script4 []byte = []byte(string(SignTransaction(sk1, tx1)) + "*" + string(SignTransaction(sk2, tx1)))

func TestVerifyTxSignature4(t *testing.T) {
	success := blockchain.verifyScripts(tx1, signature_script4, pk_script4)

	fmt.Println(success)
	if !success {
		t.Fatalf("It should pass, but it doesn't.")
	}
}

func TestGenerateP2PKHPkScript(t *testing.T) {
	pk_script := GenerateP2PKHPkScript(PK)
	signature_script := SignTransaction(SK, tx1)
	success := blockchain.verifyScripts(tx1, signature_script, pk_script)

	fmt.Println(success)
	if !success {
		t.Fatalf("It should pass, but it doesn't.")
	}
}

func TestGenerateMultisigPkScript(t *testing.T) {
	pks := []dsa.PublicKey{pk1, pk2, pk3}
	pk_script := GenerateMultisigPkScript(pks, 3, 2)
	signature_script := []byte(string(SignTransaction(sk1, tx1)) + "*" + string(SignTransaction(sk2, tx1)))
	success := blockchain.verifyScripts(tx1, signature_script, pk_script)

	fmt.Println(success)
	if !success {
		t.Fatalf("It should pass, but it doesn't.")
	}
}

func TestVerifyTransaction(t *testing.T) {
	if blockchain.verifyTransaction(tx1, false) {
		t.Fatalf("It should return false, but it returns true")
	}
}

func TestAddTransaction(t *testing.T) {
	var wallet Wallet
	wallet.Init(&blockchain)
	blockchain.init(&wallet)
	blockchain.addTransaction(tx1)
	hash1, _ := utils.GetHash(&tx1)
	tx2 := blockchain.Mempool[hash1]
	hash2, _ := utils.GetHash(&tx2)
	if hash1 != hash2 {
		t.Fatalf("Failed to add transaction tx1")
	}
}

func TestVerifyBlock(t *testing.T) {
	if blockchain.verifyBlock(sblk1, 0) {
		t.Fatalf("It should return false, but it returns true")
	}
}

func TestAddBlock(t *testing.T) {
	var sblks []message.SerializedBlock = []message.SerializedBlock{sblk1}
	blockchain.addBlock(0, sblks)
	if len(blockchain.Block) == 0 {
		t.Fatalf("It should add this block, but it doesn't.")
	}
}

func TestMine(t *testing.T) {

}
