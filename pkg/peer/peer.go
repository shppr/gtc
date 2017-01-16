package peer

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/mbags/gtc/pkg/bitfield"
	"github.com/mbags/gtc/pkg/util"
)

/*
A block is downloaded by the client when the client is interested in a peer, and that peer is not choking the client. A block is uploaded by a client when the client is not choking a peer, and that peer is interested in the client.

It is important for the client to keep its peers informed as to whether or not it is interested in them. This state information should be kept up-to-date with each peer even when the client is choked. This will allow peers to know if the client will begin downloading when it is unchoked (and vice-versa).
*/

// Peer A peer to connect to
type Peer struct {
	IP             net.IP
	Port           uint16
	Conn           net.Conn
	ID             string
	amChoking      bool
	amInterested   bool
	peerChoking    bool
	peerInterested bool
	Bitfield       bitfield.Bitfield
}

// Connect connects to a peer, handshakes, and checks for matching infohash
func (p *Peer) Connect(infoHash, peerID []byte, activate, deactivate chan<- *Peer) {
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
		p.amInterested = true
	}
	log.Printf("Connected to peer: %v", p.IP)
	p.readMessages(conn, activate, deactivate)
}

func (p *Peer) readMessages(conn net.Conn, activate, deactivate chan<- *Peer) {
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

		connectionResponseBytes = make([]byte, length-1)
		_, err = conn.Read(connectionResponseBytes)

		switch messageID {
		case 0:
			p.peerChoking = true
			log.Printf("Choked by peer %s :: %s\n", p.IP, p.ID)
			deactivate <- p
		case 1:
			p.peerChoking = false
			log.Printf("Unchoked by peer %s :: %s\n", p.IP, p.ID)
			activate <- p
		case 2:
			p.peerInterested = true
			log.Printf("Interested message by peer %s :: %s\n", p.IP, p.ID)
		case 3:
			p.peerInterested = false
			log.Printf("Not interested message by peer %s :: %s\n", p.IP, p.ID)
		case 4:
			p.Bitfield.Set(util.BytesToInt(connectionResponseBytes))
			log.Printf("Have [%d] message from peer %s :: %s\n", util.BytesToInt(connectionResponseBytes), p.IP, p.ID)
		case 5:
			p.Bitfield.Bits = connectionResponseBytes
			log.Printf("Bitfield message from peer %s :: %s\n", p.IP, p.ID)
		case 6:
			log.Printf("Request message from peer %s :: %s\n", p.IP, p.ID)
		case 7:
			log.Printf("Piece message from peer %s :: %s\n", p.IP, p.ID)
		case 8:
			log.Printf("Cancel message from peer %s :: %s\n", p.IP, p.ID)
		case 9: // for DHT later
			log.Printf("Port message from %s :: %s\n", p.IP, p.ID)
		default:
			log.Printf("Message id %d received from peer %s :: %s\n", messageID, p.IP, p.ID)
		}
	}
}

func (p *Peer) readN(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(p.Conn, buf) // look up LimitedReader or something instead later
	return buf, err
}
