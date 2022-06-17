package main

import (
	"crypto/dsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

func main() {
	var params dsa.Parameters

	if e := dsa.GenerateParameters(&params, rand.Reader, dsa.L1024N160); e != nil {
		fmt.Println(e)
	}

	var priv dsa.PrivateKey

	priv.Parameters = params
	if e := dsa.GenerateKey(&priv, rand.Reader); e != nil {
		fmt.Println(e)
	}

	pub := priv.PublicKey
	fmt.Printf("%T\n", priv)

	message := []byte("hello world")

	r, s, e := dsa.Sign(rand.Reader, &priv, message)
	if e != nil {
		fmt.Println(e)
	}

	fmt.Printf("%T\n%T\n", r, s)

	sig := r.String() + "*" + s.String()
	fmt.Println(sig)

	rands := strings.FieldsFunc(sig, split)
	fmt.Printf("%T\n", rands)
	r1 := rands[0]
	s1 := rands[1]
	println(r1, s1)

	if dsa.Verify(&pub, message, r, s) {
		fmt.Println("认证正确")
	} else {
		fmt.Println("认证失败")
	}

	fmt.Println("\n", params.P)
	fmt.Println("\n", params.Q)
	fmt.Println("\n", params.G)

	fmt.Println("\n", params)
	fmt.Println("\n", Parameters2Bytes(params))
	fmt.Println("\n", ReadParameters(Parameters2Bytes(params)))

	fmt.Println(ReadParameters(Parameters2Bytes(params)) == params)

	fmt.Println("\n", pub)
	fmt.Println("\n", PK2Bytes(pub))
	fmt.Println("\n", ReadPK(PK2Bytes(pub)))

	fmt.Println(ReadParameters(Parameters2Bytes(params)) == params)

	fmt.Println("\n", priv)
	fmt.Println("\n", SK2Bytes(priv))
	fmt.Println("\n", ReadSK(SK2Bytes(priv)))
}

func split(s rune) bool {
	if s == '*' {
		return true
	}
	return false
}

func Parameters2Bytes(parameters dsa.Parameters) (b []byte) {
	p := *parameters.P
	q := *parameters.Q
	g := *parameters.G

	var result string = p.String() + "*" + q.String() + "*" + g.String()
	return []byte(result)
}

func PK2Bytes(key dsa.PublicKey) (b []byte) {
	params := key.Parameters
	part1 := Parameters2Bytes(params)
	y := *key.Y
	var result string = string(part1) + "*" + y.String()
	return []byte(result)
}

func SK2Bytes(key dsa.PrivateKey) (b []byte) {
	pk := key.PublicKey
	part1 := PK2Bytes(pk)
	x := *key.X
	var result string = string(part1) + "*" + x.String()
	return []byte(result)
}

func ReadParameters(b []byte) dsa.Parameters {
	sk_bytes := string(b)
	param := strings.FieldsFunc(sk_bytes, split)
	p := param[0]
	q := param[1]
	g := param[2]

	big_p, _ := new(big.Int).SetString(p, 10)
	big_q, _ := new(big.Int).SetString(q, 10)
	big_g, _ := new(big.Int).SetString(g, 10)

	return dsa.Parameters{P: big_p, Q: big_q, G: big_g}
}

func ReadPK(b []byte) dsa.PublicKey {
	pk_bytes := string(b)
	param := strings.FieldsFunc(pk_bytes, split)
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

func ReadSK(b []byte) dsa.PrivateKey {
	sk_bytes := string(b)
	param := strings.FieldsFunc(sk_bytes, split)
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
