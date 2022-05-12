package main

import (
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
	Mining  bool
	// The height of blocks
	// used to examine the existence of a block
	// TODO: maintain this information
	Height map[[32]byte]int
}

func (b *BlockChain) verifyTransaction(tx message.Transaction) bool {
	// Currently, this function only verified the wallet of input >= the wallet of output
	in_count := tx.Tx_in_count
	tx_in := tx.Tx_in
	out_count := tx.Tx_out_count
	tx_out := tx.Tx_out

	wallet := int64(0)
	for i := 0; i < int(in_count); i++ {
		ID := tx_in[i].Previous_output.Hash
		index := tx_in[i].Previous_output.Index
		wallet += b.TX[ID].Tx_out[index].Value // tx[ID] or mempool[ID]
	}
	for i := uint64(0); i < out_count; i++ {
		wallet -= tx_out[i].Value
	}

	if wallet < 0 {
		return false
	}

	return true
}

func (b *BlockChain) addTransaction(tx message.Transaction) {
	// TODO add to transaction
	b.Mining = false
	txID, _ := utils.GetHash(&tx)
	b.Mempool[txID] = tx
}

func (b *BlockChain) verifyBlock(startPos int, sBlock message.SerializedBlock) bool {
	newBlock := sBlock.Header
	//newBlockHash := sBlock.HeaderHash
	newTransactions := sBlock.Txns

	lastBlockHash := b.Block[len(b.Block)-1].HeaderHash
	if newBlock.Previous_block_header_hash != lastBlockHash {
		return false
	}

	if newBlock.Merkle_root_hash != message.MakeMerkleTree(newTransactions) {
		return false
	}

}

func (b *BlockChain) addBlock(startPos int, newBlocks []message.SerializedBlock) {
	// TODO
	// len(b.block) < starPos + len(newBlocks)
	b.Mtx.Lock()
	chainLength := len(b.Block)
	newChainLength := startPos - 1 + len(newBlocks)
	if chainLength < newChainLength && chainLength >= startPos {
		var staleTransactions []message.Transaction // stale transactions
		for i := startPos; i <= chainLength-1; i++ {
			transactions := b.Block[i].Txns
			for j := 0; j < len(transactions); j++ {
				staleTransactions = append(staleTransactions, transactions[j])
			}
		}

		for i := 0; i < len(staleTransactions); i++ { // roll back
			transaction := staleTransactions[i]
			hash, _ := utils.GetHash(transaction)
			delete(b.TX, hash)
			b.Mempool[hash] = transaction
		}

		var validTransactions []message.Transaction
		for i := 0; i < len(newBlocks); i++ {
			newTransactions := newBlocks[i].Txns
			for j := 0; j < len(newTransactions); j++ {
				validTransactions = append(validTransactions, newTransactions[j])
			}
		}

		for i := 0; i < len(validTransactions); i++ {
			transaction := validTransactions[i]
			hash, _ := utils.GetHash(transaction)
			b.TX[hash] = transaction
			delete(b.Mempool, hash)
		}

		b.Block = b.Block[:startPos]
		for i := 0; i < len(newBlocks); i++ {
			b.Block = append(b.Block, newBlocks[i])
		}
	}
	b.Mtx.Unlock()
}

func (b *BlockChain) mine(version int32, TS []message.Transaction, nBits uint32, peer Peer) {
	b.Mining = true
	previous_block_header_hash := b.Block[len(b.Block)-1].HeaderHash
	nonce := uint32(0)
	for b.Mining {
		block, err := message.CreateBlock(version, previous_block_header_hash, TS, nBits, nonce)
		if err == nil {
			//newBlock := []message.Block{block}
			var serializedBlock message.SerializedBlock
			serializedBlock.Header = block
			serializedBlock.HeaderHash, _ = utils.GetHash(block)
			serializedBlock.Txns = TS

			var newBlock []message.SerializedBlock
			newBlock = append(newBlock, serializedBlock)
			b.addBlock(len(b.Block), newBlock)

			peer.BroadcastBlock()
			break
		}
		nonce++
	}
}
