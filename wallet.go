package main

import (
	"crypto/dsa"
	"log"
	"sync"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

type Account struct {
	key dsa.PrivateKey
	UTXO map[message.Outpoint]void
}

type Wallet struct {
	mtx *sync.Mutex
	blockchain *BlockChain
	Accounts map[string]*Account
	Pubkey map[string]dsa.PublicKey
	keyowner map[dsa.PublicKey]*Account // only logs those we own privkey
}

func (w *Wallet) Init(b *BlockChain) {
	w.blockchain = b
	w.mtx = &b.Mtx
	w.Accounts = make(map[string]*Account)
	w.Pubkey = make(map[string]dsa.PublicKey)
	w.keyowner = make(map[dsa.PublicKey]*Account)
}

func (w *Wallet) AddPubKey(name string, pub dsa.PublicKey) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	ret, ok := w.Pubkey[name]
	if ok {
		if ret == pub {
			return
		}
		log.Printf("[ERROR] Name %v has an existing pubkey: %v\n", name, string(PK2Bytes(pub)))
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
		log.Printf("[ERROR] Name %v has an existing privkey: %v\n", name, string(SK2Bytes(prv)))
		return
	}
	if pub, ok := w.Pubkey[name]; ok && pub != prv.PublicKey {
		log.Printf("[ERROR] Name %v has an existing pubkey: %v\n", name, string(PK2Bytes(pub)))
	}
	var ac Account
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
	w.keyowner[prv.PublicKey] = &ac
}

func (w *Wallet) OnTX(tx *message.Transaction) { // WARN: no lock!
	hash, _ := utils.GetHash(tx)
	for i, o := range tx.Tx_out {
		pk_script := Bytes2PK(o.Pk_script)
		acc, ok := w.keyowner[pk_script]
		if ok {
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
