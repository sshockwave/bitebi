package main

import (
	"crypto/dsa"
	"log"
	"sync"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

type Account struct {
	name string
	key dsa.PrivateKey
	UTXO map[message.Outpoint]void
}

type Wallet struct {
	mtx *sync.Mutex
	blockchain *BlockChain
	Accounts map[string]*Account
	Pubkey map[string]dsa.PublicKey
	keyowner map[string]*Account // only logs those we own privkey
}

func (w *Wallet) Init(b *BlockChain) {
	w.blockchain = b
	w.mtx = &b.Mtx
	w.Accounts = make(map[string]*Account)
	w.Pubkey = make(map[string]dsa.PublicKey)
	w.keyowner = make(map[string]*Account)
}

func (w *Wallet) AddPubKey(name string, pub dsa.PublicKey) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	ret, ok := w.Pubkey[name]
	if ok {
		if ret == pub {
			return
		}
		log.Printf("[ERROR] Name %v has an existing pubkey: %v", name, string(PK2Bytes(pub)))
		return
	}
	w.Pubkey[name] = pub
}

func (w *Wallet) AddPrivKey(name string, prv dsa.PrivateKey) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if ret, ok := w.Accounts[name]; ok {
		if ret.key == prv {
			return
		}
		log.Printf("[ERROR] Name %v has an existing privkey: %v", name, string(SK2Bytes(prv)))
		return
	}
	if pub, ok := w.Pubkey[name]; ok && pub != prv.PublicKey {
		log.Printf("[ERROR] Name %v has an existing pubkey: %v", name, string(PK2Bytes(pub)))
	}
	var ac Account
	ac.name = name
	ac.key = prv
	ac.UTXO = make(map[message.Outpoint]void)
	self_script := string(PK2Bytes(ac.key.PublicKey))
	for outPoint, val := range w.blockchain.UTXO {
		if !val {
			continue
		}
		pk_script := w.blockchain.TX[outPoint.Hash].Tx_out[outPoint.Index].Pk_script
		if string(pk_script) == self_script {
			ac.UTXO[outPoint] = void_null
		}
	}
	w.Accounts[name] = &ac
	w.Pubkey[name] = prv.PublicKey
	w.keyowner[string(PK2Bytes(prv.PublicKey))] = &ac
}

func (w *Wallet) OnTX(tx *message.Transaction) { // WARN: no lock!
	hash, _ := utils.GetHash(tx)
	for i, o := range tx.Tx_out {
		pk_script := FindAccountFromPkScript("P2PKH", o.Pk_script)
		if len(pk_script) == 0 {
			continue
		}
		acc, ok := w.keyowner[string(PK2Bytes(pk_script[0]))]
		if ok {
			log.Printf("[INFO] New balance for %v: %v", acc.name, o.Value)
			acc.UTXO[message.NewOutPoint(hash, uint32(i))] = void_null
		}
	}
}

func (w *Wallet) MakeTxIn(name string, value int64) (sum int64, o []message.Outpoint) {
	o = make([]message.Outpoint, 0)
	w.mtx.Lock()
	defer w.mtx.Unlock()
	acc, ok := w.Accounts[name]
	if !ok {
		return
	}
	var remove_list []message.Outpoint
	for oput, _ := range acc.UTXO {
		if ok1, ok2 := w.blockchain.UTXO[oput]; !ok1 || !ok2 {
			remove_list = append(remove_list, oput)
			continue
		}
		o = append(o, oput)
		sum += w.blockchain.TX[oput.Hash].Tx_out[oput.Index].Value
		if sum >= value {
			break
		}
	}
	for _, o := range remove_list {
		delete(acc.UTXO, o)
	}
	return
}

func (w *Wallet) RemoveUTXO(name string, remove_list []message.Outpoint) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	acc, ok := w.Accounts[name]
	if !ok {
		return
	}
	for _, o := range remove_list {
		delete(acc.UTXO, o)
	}
}

func (w *Wallet) GetSK(name string) (sk dsa.PrivateKey, ok bool) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	acc, ok := w.Accounts[name]
	if !ok {
		ok = false
		return
	}
	return acc.key, true
}
