package p2p

// https://developer.bitcoin.org/reference/p2p_networking.html

type NetConfig struct {
    DefaultPort int
    StartString [4]byte
    MaxNBits uint32
}

// Constants taken from
// https://github.com/bitcoin/bitcoin/blob/master/src/chainparams.cpp
func GetMainnet() NetConfig {
    return NetConfig{8333, [4]byte{0xf9, 0xbe, 0xb4, 0xd9}, 0x1d00ffff};
}
func GetTestnet() NetConfig {
    return NetConfig{18333, [4]byte{0x0b, 0x11, 0x09, 0x07}, 0x1d00ffff};
}
func GetRegtest() NetConfig {
    return NetConfig{18444, [4]byte{0xfa, 0xbf, 0xb5, 0xda}, 0x207fffff};
}
func GetBitebinet() NetConfig {
    return NetConfig{8333, [4]byte{0xf9, 0xbe, 0xb4, 0xd9}, 0x1E08ffff};
}
