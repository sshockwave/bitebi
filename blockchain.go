package main

import (
	"bytes"
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
	Height     map[[32]byte]int
	UTXO       map[message.Outpoint]bool
	ClientName string
}

func (b *BlockChain) verifyTxIn(in message.TxIn) (bool, int64) { // The first value returns whether it's valid, the second value returns its money
	previous_output := in.Previous_output
	signature_scripts := in.Signature_script
	_, ok := b.UTXO[previous_output]
	if !ok {
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

func (b *BlockChain) verifyTransaction(tx message.Transaction) bool {
	// Currently, this function only verified the wallet of input >= the wallet of output
	in_count := len(tx.Tx_in)
	tx_in := tx.Tx_in
	out_count := len(tx.Tx_out)
	tx_out := tx.Tx_out

	wallet := int64(0) // wallet varification
	for i := 0; i < in_count; i++ {
		/*
			ID := tx_in[i].Previous_output.Hash
			index := tx_in[i].Previous_output.Index
			transaction, ok := b.TX[ID]
			if !ok {
				return false
			} else {
				wallet += transaction.Tx_out[index].Value
			}*/
		valid, money := b.verifyTxIn(tx_in[i])
		if !valid {
			return false
		} else {
			wallet += money
		}
	}
	for i := 0; i < out_count; i++ {
		wallet -= tx_out[i].Value
	}
	if wallet < 0 {
		return false
	}

	for i := 0; i < len(tx_in); i++ { // input verification
		_, ok := b.UTXO[tx_in[i].Previous_output]
		if !ok {
			return false
		}
	}

	return true
}

func (b *BlockChain) addTransaction(tx message.Transaction) {
	// TODO add to transaction
	b.Mining = false
	txID, _ := utils.GetHash(&tx)
	b.Mempool[txID] = tx

	for i := 0; i < len(tx.Tx_in); i++ {
		delete(b.UTXO, tx.Tx_in[i].Previous_output)
	}
	for i := 0; i < len(tx.Tx_out); i++ {
		hash, _ := utils.GetHash(&tx)
		outPoint := message.NewOutPoint(hash, uint32(i))
		b.UTXO[outPoint] = true
	}
}

func (b *BlockChain) verifyBlock(startPos int, sBlock message.SerializedBlock) bool {
	newBlock := sBlock.Header
	//newBlockHash := sBlock.HeaderHash
	newTransactions := sBlock.Txns

	if startPos > len(b.Block) {
		return false
	}

	if startPos >= 1 {
		lastBlockHash := b.Block[startPos-1].HeaderHash
		if newBlock.Previous_block_header_hash != lastBlockHash { // previous_hash_verification
			return false
		}
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
	// len(b.block) < starPos + len(newBlocks)
	b.Mtx.Lock()
	chainLength := len(b.Block)
	newChainLength := startPos + len(newBlocks)
	if chainLength < newChainLength && chainLength >= startPos {
		accepted = true
		var staleTransactions []message.Transaction // stale transactions
		for i := startPos; i <= chainLength-1; i++ {
			transactions := b.Block[i].Txns
			for j := 0; j < len(transactions); j++ {
				staleTransactions = append(staleTransactions, transactions[j])
			}
		}

		for i := 0; i < len(staleTransactions); i++ { // roll back
			transaction := staleTransactions[i]
			hash, _ := utils.GetHash(&transaction)
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
			hash, _ := utils.GetHash(&transaction)
			b.TX[hash] = transaction
			delete(b.Mempool, hash)
		}

		b.Block = b.Block[:startPos]
		for i := 0; i < len(newBlocks); i++ {
			b.Block = append(b.Block, newBlocks[i])
		}
	}
	b.Mtx.Unlock()
	return accepted
}

func (b *BlockChain) mine(version int32, nBits uint32, peer *Peer) {
	b.Mining = true
	previous_block_header_hash := b.Block[len(b.Block)-1].HeaderHash

	var rewardTransaction message.Transaction = message.Transaction{
		Version: 0,
		Tx_in:   []message.TxIn{},
		Tx_out: []message.TxOut{
			{
				Value:     1, // How many bitcoins to use for reward?
				Pk_script: []byte(b.ClientName),
			},
		},
		Lock_time: 0,
	}

	var TS = []message.Transaction{rewardTransaction}
	for _, value := range b.Mempool {
		TS = append(TS, value)
	}

	nonce := uint32(0)
	for b.Mining {
		block, err := message.CreateBlock(version, previous_block_header_hash, TS, nBits, nonce)
		if err == nil {
			//newBlock := []message.Block{block}
			var serializedBlock message.SerializedBlock
			serializedBlock.Header = block
			serializedBlock.HeaderHash, _ = utils.GetHash(&block)
			serializedBlock.Txns = TS

			var newBlock []message.SerializedBlock
			newBlock = append(newBlock, serializedBlock)
			b.addBlock(len(b.Block), newBlock)

			peer.BroadcastBlock(serializedBlock)
			break
		}
		nonce++
	}
}
