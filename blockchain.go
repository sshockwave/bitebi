package main

import (
	"github.com/sshockwave/bitebi/message"
)

type BlockChain struct {
	// All blocks
	block []message.Block
	// All known transactions
	tx map[[32]byte]message.Transaction
	// The transactions that have not been added to a block
	mempool map[[32]byte]message.Transaction
	mining bool
}

func (b *BlockChain) addTransaction(tx message.Transaction) {
	// TODO add to transaction
}

func (b *BlockChain) addBlock(starPos int, newBlocks []message.Block) {
	// TODO 
	// len(b.block) < starPos + len(newBlocks)
}
