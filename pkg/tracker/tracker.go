package tracker

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/jackpal/bencode-go"
	metainfo "github.com/mbags/gtc/pkg/metainfo"
	peer "github.com/mbags/gtc/pkg/peer"
	"github.com/mbags/gtc/pkg/util"
)

const (
	PeerID = "-TR2920-" // transmission 2.920 :~)
)

func FindPeers(m *metainfo.MetaInfo) []*peer.Peer {
	reqURL := ""
	if m.Announce == "" {
	found:
		for _, list := range m.AnnounceList {
			for _, tracker := range list {
				if tracker[:4] == "http" {
					reqURL += tracker
					break found
				}
			}
		}
	} else {
		reqURL += m.Announce
	}

	reqURL += fmt.Sprintf("?info_hash=%s", url.QueryEscape(string(m.InfoHash[:])))
	reqURL += fmt.Sprintf("&peer_id=%s", PeerID+url.QueryEscape(util.SessionID(12)))
	reqURL += "&uploaded=0&downloaded=0&port=6881"
	if l := len(m.Files); l == 1 {
		reqURL += fmt.Sprintf("&left=%v", m.Files[0].Length)
	} else if l > 1 {
		len := int64(0)
		for _, f := range m.Files {
			len += f.Length
		}
		reqURL += fmt.Sprintf("&left=%v", len)
	}
	reqURL += fmt.Sprintf("&compact=1")
	res, err := http.Get(reqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	dict, err := bencode.Decode(res.Body)
	// make this conditional on successful type assertion
	peers := dict.(map[string]interface{})["peers"]
	peerList, err := GetPeerList(peers)
	if err != nil {
		log.Printf("[tracker] couldn't get peer list: %v", err)
	}
	return peerList
}

// Find peers from the tracker announce response
// p can be a string (&compact=1) or later a dictionary
func GetPeerList(p interface{}) ([]*peer.Peer, error) {
	pl := []*peer.Peer{}
	switch peers := p.(type) {
	case string: // binary string format
		pbuf := []byte(peers)
		length := len(pbuf)
		for i := 0; i < length; i += 6 {
			peer := &peer.Peer{}
			peer.IP = net.IPv4(pbuf[i], pbuf[i+1], pbuf[i+2], pbuf[i+3])
			peer.Port = binary.BigEndian.Uint16([]byte(peers[i+4 : i+6]))
			pl = append(pl, peer)
		}
	case []interface{}: // []dict format
		//doesnt happen because we only support &compact=1 anyway for now

	}
	return pl, nil
}
