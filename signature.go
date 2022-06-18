package main

import (
	"crypto/dsa"
	"crypto/rand"
	"log"
	"math/big"
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

func GeneratePkScript(txType string, pk dsa.PublicKey) []byte {
	if txType == "P2PKH" {
		pk_script := string(PK2Bytes(pk)) + "*" + "OP CHECKSIG"
		return []byte(pk_script)
	} else if txType == "P2SH" {

	}
	panic("Unsupported PKScript type")
}

func FindAccountFromPkScript(txType string, pk_script []byte) (key dsa.PublicKey) {
	if txType == "P2PKH" {
		operations := strings.FieldsFunc(string(pk_script), split)
		if len(operations) < 1 {
			return
		}
		return Bytes2PK([]byte(operations[0]))
	} else if txType == "P2SH" {

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

func Bytes2PK(b []byte) (pub dsa.PublicKey) {
	pk_bytes := string(b)
	param := strings.FieldsFunc(pk_bytes, splitKeys)
	if len(param) < 4 {
		return
	}
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
