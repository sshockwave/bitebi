package main

import (
	"sync"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

type BlockChain struct {
	// All blocks
	Block []message.SerializedBlock
	Mtx sync.Mutex
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

/*func (b *BlockChain) verifyBlock(block Block, TS []Transaction) bool {
	if block.merkle_root_hash != message.makeMerkleTree(TS) {
		return false
	}
}*/

func (b *BlockChain) addBlock(starPos int, newBlocks []message.Block) {
	// TODO
	// len(b.block) < starPos + len(newBlocks)
	b.Mtx.Lock()

	b.Mtx.Unlock()
}

func (b *BlockChain) mine(version int32, TS []message.Transaction, nBits uint32, peer Peer) {
	b.Mining = true
	previous_block_header_hash := b.Block[len(b.Block)-1].Previous_block_header_hash
	nonce := uint32(0)
	for b.Mining {
		block, err := message.CreateBlock(version, previous_block_header_hash, TS, nBits, nonce)
		if err == nil {
			newBlock := []message.Block{block}
			b.addBlock(len(b.Block), newBlock)

			peer.broadcastBlock()
			break
		}
		nonce++
	}
}
