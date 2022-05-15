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

func (o *Orphans) Init(chain *BlockChain) {
	o.Chain = chain
	o.nodes = make(map[[32]byte]*orphanNode)
}

func NewOrphanNode() *orphanNode {
	var o orphanNode
	o.successors = make(map[[32]byte]void)
	return &o
}

func (o *Orphans) AddBlock(blk *message.SerializedBlock) {
	hash, _ := utils.GetHash(&blk.Header)
	o.Chain.Mtx.Lock()
	defer o.Chain.Mtx.Unlock()
	node := o.nodes[hash]
	if node == nil {
		node = NewOrphanNode()
		o.nodes[hash] = node
	}
	node.blk = blk
	prev_hash := node.blk.Header.Previous_block_header_hash
	par := o.nodes[prev_hash]
	if par == nil {
		par = NewOrphanNode()
		o.nodes[prev_hash] = par
	}
	par.successors[hash] = void_null
}

func (o *Orphans) RemoveBlock(hash [32]byte, delay uint64) {
	if delay != 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}
	o.Chain.Mtx.Lock()
	defer o.Chain.Mtx.Unlock()
	node, ok := o.nodes[hash]
	if !ok {
		return
	}
	if node.blk == nil {
		if len(node.successors) == 0 {
			log.Fatalln("[FATAL] node.blk == nil should indicate that the block is only required by others")
		}
		return
	}
	prev_hash := node.blk.Header.Previous_block_header_hash
	node.blk = nil
	if len(node.successors) == 0 {
		delete(o.nodes, hash)
	}
	par := o.nodes[prev_hash]
	delete(par.successors, hash)
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
	for i, j := 0, len(stk) - 1; i < j; i, j = i + 1, j - 1 {
		stk[i], stk[j] = stk[j], stk[i]
	}
	return stk
}
