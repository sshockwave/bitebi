package main

import (
	"github.com/sshockwave/bitebi/message"
	"sync"
)

type BlockChain struct {
	// All blocks
	block []message.Block
	mutex sync.Mutex
	// All known transactions
	tx map[[32]byte]message.Transaction
	// The transactions that have not been added to a block
	mempool map[[32]byte]message.Transaction
	mining  bool
}

func (b *BlockChain) verifyTransaction(tx message.Transaction) bool {
	in_count := tx.Tx_in_count
	tx_in := tx.Tx_in
	out_count := tx.Tx_out_count
	tx_out := tx.Tx_out

	wallet := int(0)
	for i := 0; i < int(in_count); i++ {
		ID := tx_in[i].Previous_output.hash
		index := tx_in[i].Previous_output.index
		wallet += b.tx[ID].tx_out[index].value // tx[ID] or mempool[ID]
	}
	for i := 0; i < out_count; i++ {
		wallet -= tx_out[i].value
	}

	if wallet < 0 {
		return false
	}

	return true
}

func (b *BlockChain) addTransaction(tx message.Transaction) {
	// TODO add to transaction
	b.mining = false
	txID := message.GetHash(tx)
	b.mempool[txID] = tx
}

/*func (b *BlockChain) verifyBlock(block Block, TS []Transaction) bool {
	if block.merkle_root_hash != message.makeMerkleTree(TS) {
		return false
	}
}*/

func (b *BlockChain) addBlock(starPos int, newBlocks []message.Block) {
	// TODO
	// len(b.block) < starPos + len(newBlocks)
	b.mtx.Lock()

	b.mtx.Unlock()
}

func (b *BlockChain) mine(version int32, TS []message.Transaction, nBits int32, peer Peer) {
	b.mining = true
	previous_block_header_hash := b.block[len(b.block)-1].previous_block_header_hash
	nonce := uint32(0)
	var block message.Block
	for b.mining {
		block, err := message.CreateBlock(version, previous_block_header_hash, TS, nBits, nonce)
		if err == nil {
			newBlock := []message.Block(block)
			b.addBlock(len(b.block), newBlock)

			peer.broadcastBlock()
			break
		}
		nonce++
	}
}
