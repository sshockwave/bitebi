package main

import (
	"crypto/dsa"
	"fmt"
	"github.com/sshockwave/bitebi/message"
	"testing"
)

var tx2 message.Transaction = message.Transaction{
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

var SK dsa.PrivateKey = GenPrivKey()
var PK dsa.PublicKey = SK.PublicKey

func TestSignSignature(t *testing.T) {
	success := VerifyTxSignature(PK, []byte(SignTransaction(SK, tx2)), tx2)
	//success := Verify(pk, signature_script2, []byte("007"))

	fmt.Println(success)
	if !success {
		t.Fatalf("It should pass, but it doesn't.")
	}
}
