package peer

import (
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

	// need choke and interest states here
}

// Connect connects to a peer, handshakes, and checks for matching infohash
func (p *Peer) Connect(infoHash, peerID []byte) error {
	buf := bytes.Buffer{}
	buf.WriteByte(19)
	buf.WriteString("BitTorrent protocol\x00\x00\x00\x00\x00\x00\x00\x00")
	buf.Write(infoHash)
	buf.Write(peerID)
	handshake := buf.Bytes()

	conn, err := net.Dial("tcp", p.IP.String()+":"+strconv.Itoa(int(p.Port)))
	if err != nil {
		return err
	}

	// do handshake

	if _, err := conn.Write(handshake); err != nil {
		log.Printf("Send handshake failed w/ : %v\n", p.IP)
		return err
	}

	p.Conn = conn

	res, err := p.readN(68)
	if err != nil {
		log.Printf("Couldnt get handshake response from: %v\n", p.IP)
		return err
	}

	// TODO cut up the handshake and set fields in peer

	if peerInfoHash := string(res[28:48]); peerInfoHash == string(infoHash) {
		p.ID = string(res[48:])
	}
	log.Printf("Connected to peer: %v", p.IP)
	return nil
}

func (p *Peer) readN(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(p.Conn, buf) // look up LimitedReader or something instead later
	return buf, err
}
