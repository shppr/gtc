package peer

import (
    "encoding/binary"
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
)

// Peer A peer to connect to
type Peer struct {
	IP   net.IP
	Port uint16
	Conn net.Conn
	ID   string
    am_choking bool
    am_interested bool
    peer_choking bool
    peer_interested bool
}

// Connect connects to a peer, handshakes, and checks for matching infohash
func (p *Peer) Connect(infoHash, peerID []byte) {
	buf := bytes.Buffer{}
	buf.WriteByte(19)
	buf.WriteString("BitTorrent protocol\x00\x00\x00\x00\x00\x00\x00\x00")
	buf.Write(infoHash)
	buf.Write(peerID)
	handshake := buf.Bytes()

    log.Printf("Connecting to %s\n", p.IP)
	conn, err := net.Dial("tcp", p.IP.String()+":"+strconv.Itoa(int(p.Port)))
	if err != nil {
        log.Printf("Couldn't connect to %s\n", p.IP)
		return
	}

	// do handshake

    log.Printf("Sending handshake to %s\n", p.IP)
	if _, err := conn.Write(handshake); err != nil {
		log.Printf("Send handshake failed w/ : %v\n", p.IP)
		return
	}

	p.Conn = conn

	res, err := p.readN(68)
	if err != nil {
		log.Printf("Couldnt get handshake response from: %v\n", p.IP)
		return
	}

	// TODO cut up the handshake and set fields in peer

	if peerInfoHash := string(res[28:48]); peerInfoHash == string(infoHash) {
		p.ID = string(res[48:])
	}
	log.Printf("Connected to peer: %v", p.IP)
    p.readMessages(conn)
}

func (p *Peer) readMessages(conn net.Conn) {
    var connectionResponseBytes = make([]byte, 4)
    var messageID byte
    var length uint32
    var err error

    for {
        connectionResponseBytes = make([]byte, 4)
        _, err = conn.Read(connectionResponseBytes)
        if err != nil {
            log.Println("Error receiving response")
        }
        connectionResponse := bytes.NewBuffer(connectionResponseBytes)
        err = binary.Read(connectionResponse, binary.BigEndian, &length)
        if err != nil {
            return
        }
        if length == 0 {
            log.Printf("Keep-alive message from peer %s\n", p.IP)
            continue
        }
        connectionResponseBytes = make([]byte, 1)
        _, err = conn.Read(connectionResponseBytes)
        connectionResponse = bytes.NewBuffer(connectionResponseBytes)
        err = binary.Read(connectionResponse, binary.BigEndian, &messageID)
        if err != nil {
            log.Printf("Error receiving message id from peer %s\n", p.IP)
            return
        }

        connectionResponseBytes = make([]byte, length - 1)
        _, err = conn.Read(connectionResponseBytes)

        switch messageID {
            case 0:
                p.peer_choking = true
                log.Printf("Chocked by peer %s\n", p.IP)
            case 1:
                p.peer_choking = false
                log.Printf("Unchocked by peer %s\n", p.IP)
            case 2:
                p.peer_interested = true
                log.Printf("Interested message by peer %s\n", p.IP)
            case 3:
                p.peer_interested = false
                log.Printf("Not interested message by peer %s\n", p.IP)
            case 4:
                log.Printf("Have message from peer %s\n", p.IP)
            case 5:
                log.Printf("Bitfield message from peer %s\n", p.IP)
            case 6:
                log.Printf("Request message from peer %s\n", p.IP)
            case 7:
                log.Printf("Piece message from peer %s\n", p.IP)
            case 8:
                log.Printf("Cancel message from peer %s\n", p.IP)
            default:
                log.Printf("Message id %d received from peer %s\n", messageID, p.IP)
        }
    }
}

func (p *Peer) readN(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(p.Conn, buf) // look up LimitedReader or something instead later
	return buf, err
}
