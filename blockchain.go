package main

import (
	"bytes"
	"encoding/hex"
	"log"
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
	// The height of blocks
	// used to examine the existence of a block
	Height map[[32]byte]int
	UTXO   map[message.Outpoint]bool
}

func (b *BlockChain) verifyTxIn(in message.TxIn) (bool, int64) { // The first value returns whether it's valid, the second value returns its money
	previous_output := in.Previous_output
	signature_scripts := in.Signature_script
	used, ok := b.UTXO[previous_output]
	if !used || !ok {
		return false, int64(0)
	} else {
		hash := previous_output.Hash
		index := previous_output.Index
		txOut := b.TX[hash].Tx_out[index]
		if bytes.Compare(txOut.Pk_script, signature_scripts) != 0 {
			return false, int64(0)
		} else {
			return true, txOut.Value
		}
	}
}

func (b *BlockChain) init() {
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
}

// Verify if this tx is valid without examining the links and states
func (b *BlockChain) verifyTransaction(tx message.Transaction, isCoinbase bool) bool {
	wallet := int64(0) // wallet varification
	for i := 0; i < len(tx.Tx_in); i++ {
		valid, money := b.verifyTxIn(tx.Tx_in[i])
		if !valid {
			return false
		} else {
			wallet += money
		}
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
		if !ok || !ans {
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
		if bytes.Compare(newBlocks[i + 1].Header.Previous_block_header_hash[:], newBlocks[i].HeaderHash[:]) != 0 {
			return false
		}
	}
	// verify block content
	for i := range newBlocks {
		if !b.verifyBlock(newBlocks[i], startPos + i) {
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
			log.Println("[INFO] New block at height ", height, ": ", hex.EncodeToString(newBlocks[i].HeaderHash[:]))
		}
	}
	go b.refreshMining()
	return true
}

func (b *BlockChain) mine(version int32, nBits uint32, peer *Peer, Pk_script []byte) {
	var rewardTransaction message.Transaction = message.Transaction{
		Version: 0,
		Tx_in:   []message.TxIn{},
		Tx_out: []message.TxOut{
			{
				Value:     CoinBaseReward, // How many bitcoins to use for reward?
				Pk_script: Pk_script,
			},
		},
		Lock_time: 0,
	}
	var TS []message.Transaction
	ver := -1
	var block message.Block
	for {
		if ver < b.MineVersion {
			failed := make([][32]byte, 0)
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
			b.addTransaction(rewardTransaction)
			for hash, value := range b.Mempool {
				if b.confirmTransaction(value, false) {
					TS = append(TS, value)
				} else {
					failed = append(failed, hash)
				}
			}
			// rollback
			for _, value := range TS[1:] {
				b.cancelTransaction(value, false)
			}
			b.delTransaction(rewardTransaction)
			// useless tx
			for _, hash := range failed {
				delete(b.Mempool, hash)
			}
			previous_block_header_hash := b.Block[height-1].HeaderHash
			b.Mtx.Unlock()
			block = message.CreateBlock(version, previous_block_header_hash, TS, nBits, 0)
			block.Nonce = 0
		}
		hash, err := utils.GetHash(&block)
		if err == nil && utils.HasValidHash(hash, nBits) {
			log.Println("[INFO] (", string(Pk_script), ") A new block is successfully mined!!!!")
			var serializedBlock message.SerializedBlock
			serializedBlock, err = message.CreateSerialBlock(block, TS)
			if err != nil {
				log.Fatalf("[FATAL] Block creation failed")
				continue
			}
			b.addBlock(len(b.Block), []message.SerializedBlock{serializedBlock})
			peer.BroadcastBlock(serializedBlock)
		}
		block.Nonce++
	}
}

func (b *BlockChain) ResumeMining() {
	b.MineBarrier.Unlock()
}

func (b *BlockChain) PauseMining() {
	b.MineBarrier.Lock()
	b.Mtx.Lock()
	defer b.Mtx.Unlock()
	b.MineVersion++
}

func (b *BlockChain) refreshMining() {
	b.PauseMining()
	b.ResumeMining()
}
