package main

import (
	"crypto/dsa"
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/mr-tron/base58"
	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

func Sign(key dsa.PrivateKey, message []byte) (signature []byte) {
	r, s, _ := dsa.Sign(rand.Reader, &key, message)
	sig := r.String() + "#" + s.String()
	return []byte(sig)
}

func SignTransaction(key dsa.PrivateKey, transaction message.Transaction) (signature []byte) {
	txCopy := transaction
	for i := 0; i < len(txCopy.Tx_in); i++ {
		txCopy.Tx_in[i].Signature_script = nil
	}
	hash, _ := utils.GetHash(&txCopy)
	signature = Sign(key, hash[:])
	return signature
}

func Verify(key dsa.PublicKey, signature []byte, message []byte) bool {
	sig := string(signature)
	rands := strings.FieldsFunc(sig, splitSignatures)
	r := rands[0]
	s := rands[1]
	big_r, _ := new(big.Int).SetString(r, 10)
	big_s, _ := new(big.Int).SetString(s, 10)
	return dsa.Verify(&key, message, big_r, big_s)
}

func VerifyTxSignature(key dsa.PublicKey, signature []byte, transaction message.Transaction) bool {
	txCopy := transaction
	txCopy.Tx_in = make([]message.TxIn, len(transaction.Tx_in))
	for i := 0; i < len(txCopy.Tx_in); i++ {
		txCopy.Tx_in[i] = transaction.Tx_in[i]
		txCopy.Tx_in[i].Signature_script = nil
	}
	hash, _ := utils.GetHash(&txCopy)
	return Verify(key, signature, hash[:])
}

func GenerateP2PKHPkScript(pk dsa.PublicKey) []byte {
	pk_script := string(PK2Bytes(pk)) + "*" + "OP CHECKSIG"
	return []byte(pk_script)
}

func GenerateMultisigPkScript(pks []dsa.PublicKey, n int, m int) []byte {
	if m > n || len(pks) != n {
		fmt.Println("Wrong parameters.")
		return nil
	} else {
		var raw_script string = strconv.Itoa(m) + "*"
		for i := 0; i < len(pks); i++ { // Assert len(pks) == n
			raw_script += string(PK2Bytes(pks[i]))
			raw_script += "*"
		}
		raw_script += strconv.Itoa(n)
		raw_script += "*"
		raw_script += "OP CHECKMULTISIG"
		pk_script := []byte(raw_script)
		return pk_script
	}
}

func FindAccountFromPkScript(txType string, pk_script []byte) (pks []dsa.PublicKey) {
	if txType == "P2PKH" {
		operations := strings.FieldsFunc(string(pk_script), split)
		if len(operations) < 1 {
			return
		}
		pk := Bytes2PK([]byte(operations[0]))
		pks = []dsa.PublicKey{pk}
	} else if txType == "multisig" {
		operations := strings.FieldsFunc(string(pk_script), split)
		for i := 1; i <= len(operations)-3; i++ {
			pks = append(pks, Bytes2PK([]byte(operations[i])))
		}
	}
	return pks
}

func PK2Bytes(key dsa.PublicKey) (b []byte) {
	mar, _ := asn1.Marshal(key)
	return []byte(base58.Encode(mar))
}

func SK2Bytes(key dsa.PrivateKey) (b []byte) {
	mar, _ := asn1.Marshal(key)
	return []byte(base58.Encode(mar))
}

func Bytes2PK(b []byte) (pub dsa.PublicKey) {
	asn, err := base58.Decode(string(b))
	if err != nil {
		return
	}
	_, err = asn1.Unmarshal(asn, &pub)
	return
}

func Bytes2SK(b []byte) (prv dsa.PrivateKey) {
	asn, err := base58.Decode(string(b))
	if err != nil {
		return
	}
	_, err = asn1.Unmarshal(asn, &prv)
	return
}

func splitSignatures(s rune) bool {
	if s == '#' {
		return true
	}
	return false
}

func splitKeys(s rune) bool {
	if s == '$' {
		return true
	}
	return false
}

func split(s rune) bool {
	if s == '*' {
		return true
	}
	return false
}

func GenPrivKey() (priv dsa.PrivateKey) {
	// Generate private and public key
	var params dsa.Parameters
	if e := dsa.GenerateParameters(&params, rand.Reader, dsa.L1024N160); e != nil {
		log.Printf("[ERROR] Generate key parameters error!" + e.Error())
	}
	priv.Parameters = params
	if e := dsa.GenerateKey(&priv, rand.Reader); e != nil {
		log.Printf("[ERROR] Generate keys error!" + e.Error())
	}
	return
}
