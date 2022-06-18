package main

import (
	"crypto/dsa"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

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
	for i := 0; i < len(txCopy.Tx_in); i++ {
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
		raw_script += "n"
		raw_script += "*"
		raw_script += "OP CHECKMULTISIG"
		pk_script := []byte(raw_script)
		return pk_script
	}
}

func FindAccountFromPkScript(txType string, pk_script []byte) (pks []dsa.PublicKey) {
	if txType == "P2PKH" {
		operations := strings.FieldsFunc(string(pk_script), split)
		pk := Bytes2PK([]byte(operations[0]))
		pks = []dsa.PublicKey{pk}
	} else if txType == "multisig" {
		operations := strings.FieldsFunc(string(pk_script), split)
		for i := 1; i <= len(operations)-3; i++ {
			pks = append(pks, Bytes2PK([]byte(operations[i])))
		}
	}
	return
}

func Parameters2Bytes(parameters dsa.Parameters) (b []byte) {
	p := *parameters.P
	q := *parameters.Q
	g := *parameters.G

	var result string = p.String() + "$" + q.String() + "$" + g.String()
	return []byte(result)
}

func PK2Bytes(key dsa.PublicKey) (b []byte) {
	params := key.Parameters
	part1 := Parameters2Bytes(params)
	y := *key.Y
	var result string = string(part1) + "$" + y.String()
	return []byte(result)
}

func SK2Bytes(key dsa.PrivateKey) (b []byte) {
	pk := key.PublicKey
	part1 := PK2Bytes(pk)
	x := *key.X
	var result string = string(part1) + "$" + x.String()
	return []byte(result)
}

func Bytes2Parameters(b []byte) dsa.Parameters {
	param_bytes := string(b)
	param := strings.FieldsFunc(param_bytes, splitKeys)
	p := param[0]
	q := param[1]
	g := param[2]

	big_p, _ := new(big.Int).SetString(p, 10)
	big_q, _ := new(big.Int).SetString(q, 10)
	big_g, _ := new(big.Int).SetString(g, 10)

	return dsa.Parameters{P: big_p, Q: big_q, G: big_g}
}

func Bytes2PK(b []byte) dsa.PublicKey {
	pk_bytes := string(b)
	param := strings.FieldsFunc(pk_bytes, splitKeys)
	p := param[0]
	q := param[1]
	g := param[2]
	y := param[3]

	big_p, _ := new(big.Int).SetString(p, 10)
	big_q, _ := new(big.Int).SetString(q, 10)
	big_g, _ := new(big.Int).SetString(g, 10)
	big_y, _ := new(big.Int).SetString(y, 10)

	par := dsa.Parameters{P: big_p, Q: big_q, G: big_g}
	return dsa.PublicKey{Parameters: par, Y: big_y}
}

func Bytes2SK(b []byte) dsa.PrivateKey {
	sk_bytes := string(b)
	param := strings.FieldsFunc(sk_bytes, splitKeys)
	p := param[0]
	q := param[1]
	g := param[2]
	y := param[3]
	x := param[4]

	big_p, _ := new(big.Int).SetString(p, 10)
	big_q, _ := new(big.Int).SetString(q, 10)
	big_g, _ := new(big.Int).SetString(g, 10)
	big_y, _ := new(big.Int).SetString(y, 10)
	big_x, _ := new(big.Int).SetString(x, 10)

	par := dsa.Parameters{P: big_p, Q: big_q, G: big_g}
	pk := dsa.PublicKey{Parameters: par, Y: big_y}
	return dsa.PrivateKey{PublicKey: pk, X: big_x}
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
