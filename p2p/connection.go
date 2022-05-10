package p2p

import (
    "bytes"
    "net"
    "bufio"
    "fmt"
    "io"
    "crypto/sha256"
    "encoding/binary"
)

type Message struct {
    command string
    payload []byte
}

type Connection struct {
    tunnel *bufio.ReadWriter
    msg chan Message
    closed bool
}

func NewConnection(conn net.Conn, net_config NetConfig) *Connection {
    msg := make(chan Message)
    connection := Connection{
        tunnel: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
        msg: msg,
        closed: false,
    }
    go func() {
        for {
            header := make([]byte, 4 + 12 + 4 + 4)
            _, err := io.ReadFull(conn, header)
            if err != nil {
                if err == io.EOF {
                    break
                }
                fmt.Println("[ERROR] Error returned when reading header")
                break
            }
            if bytes.Compare(net_config.startString[:], header[0:4]) != 0 {
                fmt.Println("[ERROR] Invalid start string, discarding")
                break
            }
            command := string(bytes.TrimRight(header[4:16], "\x00"))
            payload_size := binary.LittleEndian.Uint32(header[16:20])
            if payload_size > net_config.maxNBits {
                fmt.Println("[ERROR] Payload too large")
                break
            }
            payload := make([]byte, payload_size)
            _, err = io.ReadFull(conn, payload)
            if err != nil {
                fmt.Println("[ERROR] Error occurred while reading payload")
                break
            }
            recv_chksum := sha256.Sum256(payload)
            recv_chksum = sha256.Sum256(recv_chksum[:])
            if bytes.Compare(recv_chksum[:4], header[20:24]) != 0 {
                fmt.Println("[ERROR] Checksum do not match in message, discarding")
                continue
            }
            msg <- Message{command, payload}
        }
        conn.Close()
        close(msg)
        connection.closed = true
    }()
    return &connection
}
