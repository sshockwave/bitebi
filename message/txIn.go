package message

type txIn struct {
	previous_output  outpoint
	script_bytes     uint
	signature_script []byte
	sequence         uint32
}

func makeSlice(tx txIn) struct {
}

type outpoint struct {
	hash  [32]byte
	index uint32
}
