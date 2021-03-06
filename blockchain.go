package main

import (
	"bytes"
	"crypto/dsa"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

const CoinBaseReward = 1

type BlockChain struct {
	// All blocks
	Block []message.SerializedBlock
	Mtx   sync.Mutex
	// All known transactions
	TX map[[32]byte]message.Transaction
	// The transactions that have not been added to a block
	Mempool     map[[32]byte]message.Transaction
	MineVersion int
	MineBarrier sync.Mutex
	MinerPaused bool
	// The height of blocks
	// used to examine the existence of a block
	Height map[[32]byte]int
	UTXO   map[message.Outpoint]bool
	Wallet *Wallet
}

func (b *BlockChain) init(w *Wallet) {
	b.TX = make(map[[32]byte]message.Transaction)
	b.Mempool = make(map[[32]byte]message.Transaction)
	b.Height = make(map[[32]byte]int)
	b.UTXO = make(map[message.Outpoint]bool)

	TS := []message.Transaction{{}}
	genesis := message.Block{
		Version:                    0,
		Previous_block_header_hash: [32]byte{},
		Merkle_root_hash:           message.MakeMerkleTree(TS),
		Time:                       0,
		NBits:                      0x03001000,
		Nonce:                      0,
	}

	genesis_full, err := message.CreateSerialBlock(genesis, TS)
	if err != nil {
		log.Fatalln(err)
	}
	b.Block = []message.SerializedBlock{genesis_full}
	b.Height[genesis_full.HeaderHash] = 0
	b.Wallet = w
}

func (b *BlockChain) verifyScripts(tx message.Transaction, signature_scripts []byte, pk_script []byte) bool {
	operations := strings.FieldsFunc(string(signature_scripts), split)
	for _, op := range strings.FieldsFunc(string(pk_script), split) {
		operations = append(operations, op)
	}

	stack := make([]string, 0)

	for i := 0; i < len(operations); i++ {
		if operations[i] == "OP CHECKSIG" {
			if len(stack) < 2 {
				return false
			}
			sig := stack[len(stack)-2]
			pk := stack[len(stack)-1]
			pass := VerifyTxSignature(Bytes2PK([]byte(pk)), []byte(sig), tx)
			if !pass {
				return false
			}
			stack = stack[:len(stack)-2]
		} else if operations[i] == "OP DUP" {
			if len(stack) < 1 {
				return false
			}
			stack = append(stack, stack[len(stack)-1])
		} else if operations[i] == "OP HASH160" {
			if len(stack) < 1 {
				return false
			}
			top := stack[len(stack)-1]
			top_hash := utils.Sha256Twice([]byte(top))
			stack = append(stack, string(top_hash[:]))
		} else if operations[i] == "OP EQUAL" {
			if len(stack) < 2 {
				return false
			}
			top1, _ := strconv.Atoi(stack[len(stack)-1])
			top2, _ := strconv.Atoi(stack[len(stack)-2])
			if top1 == top2 {
				stack[len(stack)-2] = strconv.Itoa(1)
			} else {
				stack[len(stack)-2] = strconv.Itoa(0)
			}
			stack = stack[:len(stack)-1]
		} else if operations[i] == "OP VERIFY" {
			if len(stack) < 1 {
				return false
			}
			top, _ := strconv.Atoi(stack[len(stack)-1])
			if top == 0 {
				return false
			}
			stack = stack[:len(stack)-1]
		} else if operations[i] == "OP EQUALVERIFY" {
			if len(stack) < 2 {
				return false
			}
			top1 := stack[len(stack)-1]
			top2 := stack[len(stack)-2]
			if top1 != top2 {
				return false
			}
			stack = stack[:len(stack)-2]
		} else if operations[i] == "OP CHECKMULTISIG" {
			if len(stack) < 1 {
				return false
			}
			n, _ := strconv.Atoi(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			if len(stack) < n {
				return false
			}

			var Pks []dsa.PublicKey
			for i := 0; i < n; i++ {
				pk := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				Pks = append(Pks, Bytes2PK([]byte(pk)))
			}

			if len(stack) < 1 {
				return false
			}
			m, _ := strconv.Atoi(stack[len(stack)-1])
			if len(stack) < m {
				return false
			}
			stack = stack[:len(stack)-1]
			var Sigs [][]byte
			for i := 0; i < m; i++ {
				sig := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				Sigs = append(Sigs, []byte(sig))
			}

			pk_pt := 0
			sig_pt := 0
			for pk_pt < n && sig_pt < m {
				pass := VerifyTxSignature(Pks[pk_pt], Sigs[sig_pt], tx)
				if pass {
					pk_pt++
					sig_pt++
				} else {
					pk_pt++
				}
			}

			if sig_pt < m {
				return false
			}
		} else if operations[i] == "OP RETURN" {
			return false
		} else { // Default: push into stack
			stack = append(stack, operations[i])
		}
	}

	if len(stack) == 0 {
		return true
	} else {
		return false
	}
}

// Verify if this tx is valid without examining the links and states
func (b *BlockChain) verifyTransaction(tx message.Transaction, isCoinbase bool) bool {
	wallet := int64(0) // wallet varification
	for i := 0; i < len(tx.Tx_in); i++ {
		previous_output := tx.Tx_in[i].Previous_output
		hash := previous_output.Hash
		pre_tx, ok := b.TX[hash]
		if !ok {
			return false
		}
		index := previous_output.Index
		if int(index) >= len(pre_tx.Tx_out) {
			return false
		}
		pre_out := pre_tx.Tx_out[index]
		pass := b.verifyScripts(tx, tx.Tx_in[i].Signature_script, pre_out.Pk_script)
		if !pass {
			return false
		}
		wallet += pre_out.Value
	}
	for i := 0; i < len(tx.Tx_out); i++ {
		wallet -= tx.Tx_out[i].Value
		if wallet < 0 {
			return false
		}
	}
	return true
}

func (b *BlockChain) verifyCoinbase(tx message.Transaction, height int) bool {
	if len(tx.Tx_in) != 1 {
		return false
	}
	if len(tx.Tx_in[0].Signature_script) != 4 {
		return false
	}
	hgt_bytes := tx.Tx_in[0].Signature_script[1:]
	new_height := int(hgt_bytes[0]) + (int(hgt_bytes[1]) << 8) + (int(hgt_bytes[2]) << 16)
	if new_height != height {
		return false
	}
	wallet := int64(CoinBaseReward) // wallet varification
	for i := 0; i < len(tx.Tx_out); i++ {
		wallet -= tx.Tx_out[i].Value
		if wallet < 0 {
			return false
		}
	}
	return true
}

// Only add it to known tx pool, don't do verification
func (b *BlockChain) addTransaction(tx message.Transaction) {
	txID, _ := utils.GetHash(&tx)
	if _, ok := b.TX[txID]; ok {
		// transaction already exists
		return
	}
	b.TX[txID] = tx
	b.Mempool[txID] = tx
	hash, _ := utils.GetHash(&tx)
	for i := 0; i < len(tx.Tx_out); i++ {
		outPoint := message.NewOutPoint(hash, uint32(i))
		b.UTXO[outPoint] = true
	}
	b.Wallet.OnTX(&tx)
}

func (b *BlockChain) confirmTransaction(tx message.Transaction, isCoinbase bool) bool {
	// input verification should have been done in verify
	for i := 0; i < len(tx.Tx_in) && !isCoinbase; i++ { // input verification
		ans, ok := b.UTXO[tx.Tx_in[i].Previous_output]
		if !ok || !ans {
			// failed, roll back
			for j := 0; j < i; j++ {
				b.UTXO[tx.Tx_in[j].Previous_output] = true
			}
			return false
		}
		b.UTXO[tx.Tx_in[i].Previous_output] = false
	}
	txID, _ := utils.GetHash(&tx)
	delete(b.Mempool, txID)
	return true
}

// This transaction should already be confirmed
func (b *BlockChain) cancelTransaction(tx message.Transaction, isCoinbase bool) {
	for i := 0; i < len(tx.Tx_in) && !isCoinbase; i++ {
		ans, ok := b.UTXO[tx.Tx_in[i].Previous_output]
		if !ok || ans {
			log.Fatalf("[ERROR] the transaction should have been confirmed")
		}
		b.UTXO[tx.Tx_in[i].Previous_output] = true
	}
	txID, _ := utils.GetHash(&tx)
	b.Mempool[txID] = tx
}

func (b *BlockChain) delTransaction(tx message.Transaction) {
	hash, _ := utils.GetHash(&tx)
	delete(b.TX, hash)
	delete(b.Mempool, hash)
	for i := 0; i < len(tx.Tx_out); i++ {
		outPoint := message.NewOutPoint(hash, uint32(i))
		delete(b.UTXO, outPoint)
	}
}

// Verify if this block is valid without examining the links and states
func (b *BlockChain) verifyBlock(sBlock message.SerializedBlock, height int) bool {
	newBlock := sBlock.Header
	//newBlockHash := sBlock.HeaderHash
	newTransactions := sBlock.Txns

	if !utils.HasValidHash(sBlock.HeaderHash, newBlock.NBits) {
		return false
	}

	if newBlock.Merkle_root_hash != message.MakeMerkleTree(newTransactions) { // merkleTree_hash_verification
		return false
	}

	if !b.verifyCoinbase(newTransactions[0], height) {
		return false
	}
	for _, transaction := range newTransactions[1:] {
		if b.verifyTransaction(transaction, false) == false {
			return false
		}
	}

	return true
}

func (b *BlockChain) addBlock(startPos int, newBlocks []message.SerializedBlock) (accepted bool) {
	b.Mtx.Lock()
	defer b.Mtx.Unlock()
	// add new known transactions
	// they should always be added
	// since they might be useful
	for i := range newBlocks {
		for j := range newBlocks[i].Txns {
			b.addTransaction(newBlocks[i].Txns[j])
		}
	}
	// Consensus: always use longest chain
	if !(startPos <= len(b.Block) && startPos+len(newBlocks) > len(b.Block)) {
		return false
	}
	// verify block connect hash
	if bytes.Compare(newBlocks[0].Header.Previous_block_header_hash[:], b.Block[startPos-1].HeaderHash[:]) != 0 {
		return false
	}
	for i := range newBlocks[1:] {
		if bytes.Compare(newBlocks[i+1].Header.Previous_block_header_hash[:], newBlocks[i].HeaderHash[:]) != 0 {
			return false
		}
	}
	// verify block content
	for i := range newBlocks {
		if !b.verifyBlock(newBlocks[i], startPos+i) {
			return false
		}
	}
	// Roll back current chain
	// Permanent change, needs roll back
	for _, v := range b.Block[startPos:] {
		for i := range v.Txns {
			b.cancelTransaction(v.Txns[i], i == 0)
		}
	}
	// Add new chain
	// Permanent change, needs roll back
	for i, v := range newBlocks {
		for j := range v.Txns {
			ret := b.confirmTransaction(v.Txns[j], j == 0)
			if !ret {
				// invalid transaction, roll back all
				for ; j >= 0; j-- {
					b.cancelTransaction(v.Txns[j], j == 0)
				}
				for _, v := range newBlocks[:i] {
					for j := range v.Txns {
						b.cancelTransaction(v.Txns[j], j == 0)
					}
				}
				for _, v := range b.Block[startPos:] {
					for j := range v.Txns {
						ret := b.confirmTransaction(v.Txns[j], j == 0)
						if !ret {
							log.Fatalf("[ERROR] the blockchain should have been valid")
						}
					}
				}
				return false
			}
		}
	}
	{ // Commit success!
		for _, v := range b.Block[startPos:] {
			delete(b.Height, v.HeaderHash)
		}
		b.Block = b.Block[:startPos]
		for i := 0; i < len(newBlocks); i++ {
			height := len(b.Block)
			b.Height[newBlocks[i].HeaderHash] = height
			b.Block = append(b.Block, newBlocks[i])
			log.Printf("[INFO] New block at height %v: %v", height, newBlocks[i].HexString())
		}
	}
	go b.refreshMining()
	return true
}

func (b *BlockChain) sortedMempool() (ans [][32]byte) {
	succ := make(map[[32]byte][][32]byte)
	indeg := make(map[[32]byte]int)
	for hash := range b.Mempool {
		succ[hash] = [][32]byte{}
	}
	for hash, value := range b.Mempool {
		curdeg := 0
		for _, v := range value.Tx_in {
			arr, ok := succ[v.Previous_output.Hash]
			if ok {
				curdeg += 1
				succ[v.Previous_output.Hash] = append(arr, hash)
			}
		}
		if curdeg == 0 {
			ans = append(ans, hash)
		} else {
			indeg[hash] = curdeg
		}
	}
	if ans == nil {
		return [][32]byte{}
	}
	for i := 0; i < len(ans); i++ {
		hash := ans[i]
		for _, v := range succ[hash] {
			indeg[v] -= 1
			if indeg[v] == 0 {
				delete(indeg, v)
				ans = append(ans, v)
			}
		}
	}
	if len(ans) < len(b.Mempool) {
		log.Fatalln("[FATAL] Loop detected in unconfirmed txns. This is hardly possible.")
	}
	return
}

func (b *BlockChain) mine(version int32, nBits uint32, peer *Peer, Pk_script []byte) {
	rewardTransaction := message.Transaction{
		Tx_out: []message.TxOut{
			{
				Value:     CoinBaseReward, // How many bitcoins to use for reward?
				Pk_script: Pk_script,
			},
		},
	}
	var TS []message.Transaction
	ver := -1
	var block message.Block
	for {
		if ver < b.MineVersion {
			b.MineBarrier.Lock() // sync progress
			b.MineBarrier.Unlock()
			b.Mtx.Lock()
			ver = b.MineVersion
			height := len(b.Block)
			// https://developer.bitcoin.org/reference/transactions.html?highlight=coinbase
			rewardTransaction.Tx_in = []message.TxIn{
				{
					Previous_output: message.Outpoint{Index: 0xffff},
					Signature_script: []byte{
						0x03, // number of bytes in the height
						byte(height & 255),
						byte((height >> 8) & 255),
						byte((height >> 16) & 255),
					},
				},
			}
			TS = []message.Transaction{rewardTransaction}
			for _, hash := range b.sortedMempool() {
				value := b.Mempool[hash]
				if b.verifyTransaction(value, false) && b.confirmTransaction(value, false) {
					TS = append(TS, value)
				} else {
					b.delTransaction(value)
				}
			}
			// rollback
			for _, value := range TS[1:] {
				b.cancelTransaction(value, false)
			}
			previous_block_header_hash := b.Block[height-1].HeaderHash
			b.Mtx.Unlock()
			block = message.CreateBlock(version, previous_block_header_hash, TS, nBits, 0)
			block.Nonce = 0
		}
		hash, err := utils.GetHash(&block)
		if err == nil && utils.HasValidHash(hash, nBits) {
			log.Printf("[INFO] A new block is successfully mined!!!!")
			var serializedBlock message.SerializedBlock
			serializedBlock, err = message.CreateSerialBlock(block, TS)
			if err != nil {
				log.Fatalf("[FATAL] Block creation failed")
				continue
			}
			ok := b.addBlock(len(b.Block), []message.SerializedBlock{serializedBlock})
			if !ok {
				log.Println("[WARN] A mined block is discarded.")
			}
			peer.BroadcastBlock(serializedBlock)
		}
		block.Nonce++
	}
}

func (b *BlockChain) ResumeMining() {
	b.Mtx.Lock()
	defer b.Mtx.Unlock()
	b.MinerPaused = false
	b.MineBarrier.Unlock()
}

func (b *BlockChain) PauseMining() {
	b.Mtx.Lock()
	defer b.Mtx.Unlock()
	if !b.MinerPaused {
		b.MinerPaused = true
		b.MineBarrier.Lock()
	}
	b.MineVersion++
}

func (b *BlockChain) refreshMining() {
	b.Mtx.Lock()
	defer b.Mtx.Unlock()
	paused_before := b.MinerPaused
	if !paused_before {
		b.MinerPaused = true
		b.MineBarrier.Lock()
	}
	b.MineVersion++
	if !paused_before {
		b.MinerPaused = false
		b.MineBarrier.Unlock()
	}
}
