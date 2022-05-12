package main

import (
	"log"
	"time"

	"github.com/sshockwave/bitebi/message"
	"github.com/sshockwave/bitebi/utils"
)

type orphanNode struct {
	blk *message.SerializedBlock
	successors map[[32]byte]void
}

type Orphans struct {
	Chain *BlockChain
	nodes map[[32]byte]*orphanNode
}

func (o *Orphans) AddBlock(blk *message.SerializedBlock) {
	hash, _ := utils.GetHash(&blk.Header)
	node := o.nodes[hash]
	if node == nil {
		node = new(orphanNode)
		o.nodes[hash] = node
	}
	node.blk = blk
	prev_hash := node.blk.Header.Previous_block_header_hash
	o.Chain.Mtx.Lock()
	par := o.nodes[prev_hash]
	if par == nil {
		par = new(orphanNode)
		o.nodes[prev_hash] = par
	}
	par.successors[hash] = void_null
	o.Chain.Mtx.Unlock()
}

func (o *Orphans) RemoveBlock(hash [32]byte, delay uint64) {
	if delay != 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}
	o.Chain.Mtx.Lock()
	defer o.Chain.Mtx.Unlock()
	node := o.nodes[hash]
	if node == nil {
		return
	}
	node.blk = nil
	if len(node.successors) == 0 {
		delete(o.nodes, hash)
	}
	prev_hash := node.blk.Header.Previous_block_header_hash
	par := o.nodes[prev_hash]
	delete(par.successors, node.blk.HeaderHash)
	if par.blk == nil && len(par.successors) == 0 {
		delete(o.nodes, prev_hash)
	}
}

func (o *Orphans) dfsLongChain(hash [32]byte) (stk []*message.SerializedBlock) {
	node, ok := o.nodes[hash]
	if !ok {
		log.Fatal("[FATAL] the node should have existed")
	}
	for v := range node.successors {
		tmp_stk := o.dfsLongChain(v)
		if len(tmp_stk) > len(stk) {
			stk = tmp_stk
		}
	}
	stk = append(stk, node.blk)
	return
}

func (o *Orphans) GetLongestChain(hash [32]byte) (stk []*message.SerializedBlock) {
	o.Chain.Mtx.Lock()
	defer o.Chain.Mtx.Unlock()
	node, ok := o.nodes[hash]
	if !ok {
		return
	}
	stk = o.dfsLongChain(hash)
	for {
		node, ok = o.nodes[node.blk.Header.Previous_block_header_hash]
		if !ok {
			log.Fatal("[FATAL] Assertion failure, previous node should exist when the orhpaned block exists")
		}
		if node.blk == nil {
			break
		}
		stk = append(stk, node.blk)
	}
	for i, j := 0, 0; i < j; i, j = i + 1, j - 1 {
		stk[i], stk[j] = stk[j], stk[i]
	}
	return stk
}
