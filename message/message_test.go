package message

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/sshockwave/bitebi/utils"
)

func doSerializationTest(in utils.BinaryWritable, out utils.BinaryReadable, t *testing.T) {
	b, err := utils.GetBytes(in)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	reader := utils.NewBufReader(bytes.NewBuffer(b))
	err = out.LoadBuffer(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected equal: %v = %v", in, out)
	}
}

func TestAddr(t *testing.T) {
	addrs := AddrMsg{Addrs: []NetworkIPAddress{
		{
			time: 123456,
			services: 434432,
			Ipv6: [16]byte{0x7f,0x00,0x00,0,0,3,4,2},
			Port: 8333,
		},
		{
			time: 4433,
			services: 9929292,
			Ipv6: [16]byte{0x00,0x00,0x00,0,0,3,4,0x33},
			Port: 18333,
		},
	}}
	raw_data, err := utils.GetBytes(&addrs)
	if len(raw_data) == 0 || err != nil {
		t.Fatalf("Should have converted to bytes")
	}
	var new_addrs AddrMsg
	reader := utils.NewBufReader(bytes.NewBuffer(raw_data))
	err = new_addrs.LoadBuffer(reader)
	if err != nil {
		t.Fatal()
	}
	if !reflect.DeepEqual(addrs, new_addrs) {
		t.Fatalf("Expect equal: %v = %v", addrs, new_addrs)
	}
}

var tx1 Transaction = Transaction{
	Version: 2203,
	Tx_in: []txIn{
		{
			Previous_output: Outpoint{
				Hash: [32]byte{33,22,0,11},
				Index: 12,
			},
			script_bytes: 4,
			signature_script: []byte{22,1,1,4},
		},
		{
			Previous_output: Outpoint{
				Hash: [32]byte{8,2,6,3},
				Index: 7,
			},
			script_bytes: 5,
			signature_script: []byte{1,2,3,4,5},
		},
	},
	Tx_out: []txOut{
		{
			Value: 22555343,
			Pk_script: []byte{12,3,7,5},
		},
		{
			Value: -33,
			Pk_script: []byte{0,0,3,0,4},
		},
		{
			Value: 22,
			Pk_script: []byte{},
		},
	},
	Lock_time: 332,
}
func TestTxSerializing(t *testing.T) {
	b, err := utils.GetBytes(&tx1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var tx2 Transaction
	reader := utils.NewBufReader(bytes.NewBuffer(b))
	err = tx2.LoadBuffer(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(tx1, tx2) {
		t.Fatalf("Expected equal: %v = %v", tx1, tx2)
	}
}

var blk1 Block = Block{
	Version: -232,
	Previous_block_header_hash: [32]byte{3,3,2,2,1,4},
	Merkle_root_hash: [32]byte{1,2,3,4,55,66,77},
	Time: 10101010,
	NBits: 10,
	Nonce: 0x7f7f7f7f,
}
func TestBlockHeaderSerialization(t *testing.T) {
	var blk2 Block
	doSerializationTest(&blk1, &blk2, t)
}

var tx2 Transaction = Transaction{
	Version: 330,
	Tx_in: []txIn{},
	Tx_out: []txOut{
		{
			Value: 625,
			Pk_script: []byte("Alice"),
		},
	},
}

func TestSerializedBlockSerialization(t *testing.T) {
	sb, err := CreateSerialBlock(blk1, []Transaction{tx1, tx2})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var sb2 SerializedBlock
	doSerializationTest(&sb, &sb2, t)
}

func TestGetBlockSerialization(t *testing.T) {
	gb := GetBlocksMsg{
		Version: 2290,
		BlockHeaderHashes: [][32]byte{
			{0,1,1,5},
			{0,1,1,4},
			{4,2},
		},
		StopHash: [32]byte{55,87,79},
	}
	var new_msg GetBlocksMsg
	doSerializationTest(&gb, &new_msg, t)
}

func TestInventorySerialization(t *testing.T) {
	invmsg := InvMsg{Inv: []Inventory{
		{
			Type: MSG_TX,
			Hash: [32]byte{2,3,11,2,1,6},
		},
		{
			Type: MSG_BLOCK,
			Hash: [32]byte{0,8,7,9,9,45},
		},
		{
			Type: MSG_BLOCK,
			Hash: [32]byte{0,8,7,9,9,45},
		},
	}}
	var new_msg InvMsg
	doSerializationTest(&invmsg, &new_msg, t)
}

func bytesReverse(arr []byte) {
	for i, j := 0, len(arr) - 1; i < j; i, j = i + 1, j - 1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}
func doMerkleTreeHashTest(root_str string, hash_str []string, t *testing.T) {
    root, err := hex.DecodeString(root_str)
	bytesReverse(root)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	hash_bin := make([]byte, HashL * len(hash_str))
	for i := range hash_str {
		var data []byte
		data, err = hex.DecodeString(hash_str[i])
		bytesReverse(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		copy(hash_bin[i * HashL: (i + 1) * HashL], data)
	}
	calc_res := MakeMerkleTreeFromHashes(hash_bin)
	if bytes.Compare(root, calc_res[:]) != 0 {
		t.Fatalf("Expected hash %v, found %v", root, calc_res)
	}
}
// Taken from
// https://gist.github.com/thereal1024/45bb035e580430988a34
// These are big-endian inputs and they need to be reversed
func TestMerkleHashPower(t *testing.T) {
	doMerkleTreeHashTest(
		"0b0192e318af62f8f91243948ea4c7ea9d696197e88b9401bce35ecb0a0cb59b",
		[]string{
			"04a2808134e646ba67ff83f0bc7535a008b6e154c98953f5e2c9d40429880faf",
			"b6b3ff7b4d004a788c751f3f8fc881f96c7b647ae06eb9a720bddc924e6f9147",
			"e614ebb7e059e248e1f4c440f91af5c9617394a05d72233d7acf6feb153362f1",
			"5bbc4545145126108c91689e62c1806646468c547999241f5c2883a526e015b6",
			"de56c21783d3d466c0a5a155ed909c7011879df1996d8c418dac74465ebc3564",
			"d327f96d32afdbf4238458684570189de26ba5dc300d5cd19fa1a9cdcecdb527",
			"702c3d845810f31c194e7c9ea3d2b3636f3b8b9ee71f3d93a2f36e9d1a4e9a81",
			"b320e44b0e4cbe5973b4ebdea0c63939f9cc196982e3f4d15daaa1baa16f0004",
		},
		t,
	)
}
func TestMerkleHashUneven(t *testing.T) {
	doMerkleTreeHashTest(
		"560a4d3b44e57ff78be70d29698a8f98ce11677c1a59fb9966a7cd1795c9b47b",
		[]string{
			"df70f26b6df54332ad29c08aab5e5d5560d1468311e90484ebd89f87ac6264e8",
			"2148314cd02237786abe127f23b7346df8a116a2851745cb987652a3e132fc50",
			"06c303894833eb5d639f06f95ceb2c4bd08e0ab4ae1d94cccfa54f02e9b35990",
			"90ae3d27a5215dbb8e2e1657c927f81bdb9601106a6159f5384b4cde53836f24",
			"51cfe20029ed6366e7f475a123ad84c96c54522e9ae64cb2f548811124a6f833",
			"1e856be000b0fbaa5929b887755095106f4f0d3d19f9cd9cb07ab2239c8b4b18",
			"9d6314d68d9de8250513563e02f83ffc80973ec8b7c2966835e2cbcac3320898",
			"5d6e3fc4b0c44b867b83b7d7ca365754a8bb87d93c4f365ecacc1f0109b4c99c",
			"58afcfed0a60792c3e15d8bb2bd8d59f2a968639473e575e2fc1c270fcfae910",
			"50a0e15c32c257934f75ee2fa125dd7e9a542d38b5989efc380ea2c06a299804",
			"acd706cdbe74f82040cc583e42dfc28d8603c2f7d2fe29c0d41ee2e8d78be51b",
			"c7be55d3b55bd59f1ca19d2dc3ffbe8c28917c9e27f02456872755215b4b8a1f",
			"e323fe6719e707b8deb108d3f4bcc43d9e018cf48e027b8f88941886a0744f60",
		},
		t,
	)
}
