package p2p

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/sshockwave/bitebi/utils"
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
            if bytes.Compare(net_config.StartString[:], header[0:4]) != 0 {
                fmt.Println("[ERROR] Invalid start string, discarding")
                break
            }
            command := string(bytes.TrimRight(header[4:16], "\x00"))
            payload_size := binary.LittleEndian.Uint32(header[16:20])
            if payload_size > net_config.MaxNBits {
                fmt.Println("[ERROR] Payload too large")
                break
            }
            payload := make([]byte, payload_size)
            _, err = io.ReadFull(conn, payload)
            if err != nil {
                fmt.Println("[ERROR] Error occurred while reading payload")
                break
            }
            recv_chksum := utils.Sha256Twice(payload)
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
