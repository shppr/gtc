package tracker

import (
    "bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
    "math/rand"

	"github.com/jackpal/bencode-go"
	metainfo "github.com/mbags/gtc/pkg/metainfo"
	peer "github.com/mbags/gtc/pkg/peer"
	"github.com/mbags/gtc/pkg/util"
)

const (
	PeerID = "-TR2920-" // transmission 2.920 :~)
)

func FindPeers(m *metainfo.MetaInfo) (peerList []*peer.Peer, err error) {
	if m.Announce == "" {
	found:
		for _, list := range m.AnnounceList {
			for _, tracker := range list {
                peerList, err = queryTracker(tracker, m)
                if err != nil {
                    log.Printf("Error getting peers from %s", tracker)
                    continue found
                }
                break found
			}
		}
	} else {
        peerList, err = queryTracker(m.Announce, m)
	}
	return
}

func queryTracker(tracker string, m *metainfo.MetaInfo) (peerList []*peer.Peer, err error) {
    if tracker[:3] == "udp" {
        peerList, err = queryUDPTracker(tracker, m)
	}
    if tracker[:4] == "http" {
        peerList, err = queryHTTPTracker(tracker, m)
    }
    return
}

func queryHTTPTracker(reqURL string, m *metainfo.MetaInfo) (peerList []*peer.Peer, err error) {
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
    if err != nil {
        log.Println("bencode error")
        return
    }
	// make this conditional on successful type assertion
	peers := dict.(map[string]interface{})["peers"]
	peerList, err = GetPeerList(peers)
	if err != nil {
		log.Printf("[tracker] couldn't get peer list: %v", err)
	}
	return
}

func queryUDPTracker(reqURL string, m *metainfo.MetaInfo) (pl []*peer.Peer, err error) {
    u, err := url.Parse(reqURL)
    serverAddress, err := net.ResolveUDPAddr("udp", u.Host)

    if err != nil {
        log.Println("Error parsing URL")
        return
    }
    con, err := net.DialUDP("udp", nil, serverAddress)
    if err != nil {
        return
    }
    var connectionID uint64 = 0x41727101980
    var action uint32 = 0
    transactionID := rand.Uint32()
    request := new(bytes.Buffer)

    err = binary.Write(request, binary.BigEndian, connectionID)
    if err != nil {
        return
    }

    err = binary.Write(request, binary.BigEndian, action)
    if err != nil {
        return
    }

    err = binary.Write(request, binary.BigEndian, transactionID)
    if err != nil {
        return
    }

    _, err = con.Write(request.Bytes())
    if err != nil {
        return
    }

    var connectionResponseBytes = make([]byte, 16)
    connectionResponseLen, err := con.Read(connectionResponseBytes)
    if err != nil {
        return
    }
    if connectionResponseLen != 16 {
        log.Println("Unexpected response size")
    }

    connectionResponse := bytes.NewBuffer(connectionResponseBytes)
    var connectionResponseAction uint32
    err = binary.Read(connectionResponse, binary.BigEndian, &connectionResponseAction)
    if err != nil {
        return
    }
    if connectionResponseAction != 0 {
        log.Println("Unexpected response action")
    }
    var connectionTransactionID uint32
    err = binary.Read(connectionResponse, binary.BigEndian, &connectionTransactionID)
    if err != nil {
        return
    }
    if connectionTransactionID != transactionID {
        log.Println("Unexpected transactio id")
    }
    err = binary.Read(connectionResponse, binary.BigEndian, &connectionID)
    if err != nil {
        return
    }

    return announcementRequest(con, connectionID, m)
}

func announcementRequest(con *net.UDPConn, connectionID uint64, m *metainfo.MetaInfo) (pl []*peer.Peer, err error) {
    transactionID := rand.Uint32()

    announcementRequest := new(bytes.Buffer)
    err = binary.Write(announcementRequest, binary.BigEndian, connectionID)
    if err != nil {
        return
    }

    var action uint32 = 1
    err = binary.Write(announcementRequest, binary.BigEndian, action)
    if err != nil {
        return
    }

    err = binary.Write(announcementRequest, binary.BigEndian, transactionID)
    if err != nil {
        return
    }
    err = binary.Write(announcementRequest, binary.BigEndian, m.InfoHash)
    if err != nil {
        return
    }
    err = binary.Write(announcementRequest, binary.BigEndian, []byte(PeerID+url.QueryEscape(util.SessionID(12))))
    if err != nil {
        return
    }

    var downloaded uint64 = 0
    err = binary.Write(announcementRequest, binary.BigEndian, downloaded)
    if err != nil {
        return
    }

   err = binary.Write(announcementRequest, binary.BigEndian, m.Files[0].Length)
   if err != nil {
       return
   }

   var uploaded uint64 = 0
   err = binary.Write(announcementRequest, binary.BigEndian, uploaded)
   if err != nil {
       return
   }

    var event uint32 = 2
    err = binary.Write(announcementRequest, binary.BigEndian, event)
    if err != nil {
        return
    }

    var ipAddress uint32 = 0
    err = binary.Write(announcementRequest, binary.BigEndian, ipAddress)
    if err != nil {
        return
    }

    var key uint32 = 0
    err = binary.Write(announcementRequest, binary.BigEndian, key)
    if err != nil {
        return
    }

    const peerRequestCount = 10
    var peerCount uint32 = peerRequestCount
    err = binary.Write(announcementRequest, binary.BigEndian, peerCount)
    if err != nil {
        return
    }

    var port uint16 = 6881
    err = binary.Write(announcementRequest, binary.BigEndian, port)
    if err != nil {
        return
    }

    _, err = con.Write(announcementRequest.Bytes())
    if err != nil {
        return
    }

    const minResponseLen = 20
    const peerDataLen = 6
    expectedResponse := minResponseLen + peerDataLen*peerRequestCount
    responseBytes := make([]byte, expectedResponse)

    var responseLen int
    responseLen, err = con.Read(responseBytes)
    if err != nil {
        return
    }
    if responseLen < expectedResponse {
        log.Println("Unexpected response length")
        return
    }

    response := bytes.NewBuffer(responseBytes)

    var responseAction uint32
    err = binary.Read(response, binary.BigEndian, &responseAction)
    if err != nil {
        return
    }
    if responseAction != 1 {
        log.Println("Unexpected action response")
    }

    var responseTransactionID uint32
    err = binary.Read(response, binary.BigEndian, &responseTransactionID)
    if err != nil {
        return
    }
    if transactionID != responseTransactionID {
        log.Println("Unexpected transaction id")
    }

    var interval uint32
    err = binary.Read(response, binary.BigEndian, &interval)
    if err != nil {
        return
    }

    var leechers uint32
    err = binary.Read(response, binary.BigEndian, &leechers)
    if err != nil {
        return
    }

    var seeders uint32
    err = binary.Read(response, binary.BigEndian, &seeders)
    if err != nil {
        return
    }

    peerCountResponse := (responseLen - minResponseLen) / peerDataLen
    peerDataBytes := make([]byte, peerDataLen * peerCountResponse)
    err = binary.Read(response, binary.BigEndian, &peerDataBytes)
    if err != nil {
        return
    }

    fmt.Printf("Seeders: %d\n", seeders)
    fmt.Printf("Leechers: %d\n", leechers)
    fmt.Printf("Interval: %d\n", interval)
    fmt.Printf("Peers: %d\n", peerCountResponse)

    pl, err = GetPeerList(string(peerDataBytes))
    if err != nil {
        return
    }
    return
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
		fmt.Println("peers in dict format")
	}
	return pl, nil
}
