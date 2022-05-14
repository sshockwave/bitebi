package main

import (
	"bytes"
	"log"
	"sync"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

type BlockChain struct {
	// All blocks
	Block []message.SerializedBlock
	Mtx   sync.Mutex
	// All known transactions
	TX map[[32]byte]message.Transaction
	// The transactions that have not been added to a block
	Mempool map[[32]byte]message.Transaction
	MineVersion int
	MineBarrier sync.Mutex
	// The height of blocks
	// used to examine the existence of a block
	// TODO: maintain this information
	Height     map[[32]byte]int
	UTXO       map[message.Outpoint]bool
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
	// TODO initialize genesis block
}

// Verify if this tx is valid without examining the links and states
func (b *BlockChain) verifyTransaction(tx message.Transaction) bool {
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

func (b *BlockChain) confirmTransaction(tx message.Transaction) bool {
	for i := 0; i < len(tx.Tx_in); i++ { // input verification
		ans, ok := b.UTXO[tx.Tx_in[i].Previous_output]
		if !ok || !ans || !b.verifyTransaction(tx){
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
func (b *BlockChain) cancelTransaction(tx message.Transaction) {
	for i := 0; i < len(tx.Tx_in); i++ {
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
func (b *BlockChain) verifyBlock(sBlock message.SerializedBlock) bool {
	newBlock := sBlock.Header
	//newBlockHash := sBlock.HeaderHash
	newTransactions := sBlock.Txns

	if !utils.HasValidHash(sBlock.HeaderHash, newBlock.NBits) {
		return false
	}

	if newBlock.Merkle_root_hash != message.MakeMerkleTree(newTransactions) { // merkleTree_hash_verification
		return false
	}

	for _, transaction := range newTransactions {
		if b.verifyTransaction(transaction) == false {
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
	if !(startPos <= len(b.Block) && startPos + len(newBlocks) > len(b.Block)) {
		return false
	}
	// verify block connect hash
	if bytes.Compare(newBlocks[0].Header.Previous_block_header_hash[:], b.Block[startPos - 1].HeaderHash[:]) != 0 {
		return false
	}
	for i := range newBlocks[1:] {
		if bytes.Compare(newBlocks[i].Header.Previous_block_header_hash[:], newBlocks[i-1].HeaderHash[:]) != 0 {
			return false
		}
	}
	// verify block content
	for i := range newBlocks {
		if !b.verifyBlock(newBlocks[i]) {
			return false
		}
	}
	// Roll back current chain
	// Permanent change, needs roll back
	for _, v := range b.Block[startPos:] {
		for i := range v.Txns {
			b.cancelTransaction(v.Txns[i])
		}
	}
	// Add new chain
	// Permanent change, needs roll back
	for i, v := range newBlocks {
		for j := range v.Txns {
			ret := b.confirmTransaction(v.Txns[j])
			if !ret {
				// invalid transaction, roll back all
				for ; j >= 0; j-- {
					b.cancelTransaction(v.Txns[j])
				}
				for _, v := range newBlocks[:i] {
					for j := range v.Txns {
						b.cancelTransaction(v.Txns[j])
					}
				}
				for _, v := range b.Block[startPos:] {
					for j := range v.Txns {
						ret := b.confirmTransaction(v.Txns[j])
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
			b.Height[newBlocks[i].HeaderHash] = len(b.Block)
			b.Block = append(b.Block, newBlocks[i])
		}
	}
	go b.refreshMining()
	return true
}

func (b *BlockChain) mine(version int32, nBits uint32, peer *Peer, Pk_script []byte) {
	previous_block_header_hash := b.Block[len(b.Block)-1].HeaderHash

	var rewardTransaction message.Transaction = message.Transaction{
		Version: 0,
		Tx_in:   []message.TxIn{},
		Tx_out: []message.TxOut{
			{
				Value:     1, // How many bitcoins to use for reward?
				Pk_script: Pk_script,
			},
		},
		Lock_time: 0,
	}

	var TS []message.Transaction
	ver := -1

	block := message.CreateBlock(version, previous_block_header_hash, TS, nBits, 0)
	for {
		if ver < b.MineVersion {
			b.MineBarrier.Lock() // sync progress
			b.MineBarrier.Unlock()
			TS = []message.Transaction{rewardTransaction}
			b.Mtx.Lock()
			ver = b.MineVersion
			failed := make([][32]byte, 0)
			for hash, value := range b.Mempool {
				if b.confirmTransaction(value) {
					TS = append(TS, value)
				} else {
					failed = append(failed, hash)
				}
			}
			// rollback
			for _, value := range TS {
				b.cancelTransaction(value)
			}
			// useless tx
			for _, hash := range failed {
				delete(b.Mempool, hash)
			}
			defer b.Mtx.Unlock()
			block.Nonce = 0
		}
		hash, err := utils.GetHash(&block)
		if err == nil && utils.HasValidHash(hash, nBits) {
			//newBlock := []message.Block{block}
			var serializedBlock message.SerializedBlock
			serializedBlock.Header = block
			serializedBlock.HeaderHash, _ = utils.GetHash(&block)
			serializedBlock.Txns = TS

			var newBlock []message.SerializedBlock
			newBlock = append(newBlock, serializedBlock)
			b.addBlock(len(b.Block), newBlock)

			peer.BroadcastBlock(serializedBlock)
			go b.refreshMining()
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
