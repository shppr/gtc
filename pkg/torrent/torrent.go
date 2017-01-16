package torrent

import (
	"fmt"
	"log"
	"sync"

	"github.com/mbags/gtc/pkg/metainfo"
	"github.com/mbags/gtc/pkg/peer"
	"github.com/mbags/gtc/pkg/tracker"
	"github.com/mbags/gtc/pkg/util"
)

// Torrent torrent data(MetaInfo) and connected peers
type Torrent struct {
	MetaInfo             *metainfo.MetaInfo
	Lock                 sync.Mutex
	ActivePeers          map[string]*peer.Peer
	Activate, Deactivate chan *peer.Peer
}

// New return a Torrent struct with MetaInfo and ActivePeers populated
func NewFromFilename(filename string) (*Torrent, error) {

	m, err := metainfo.NewFromFilename(filename)
	if err != nil {
		log.Fatalf("Couldnt create metainfo for %v", filename)
	}
	// pretty print the parsed .torrent
	fmt.Println(m)
	// fetch & print initial peers
	peerList, err := tracker.FindPeers(m)
	if err != nil {
		log.Fatalf("Couldn't get the peers list")
		return nil, err
	}
	t := &Torrent{m, sync.Mutex{}, make(map[string]*peer.Peer), make(chan *peer.Peer), make(chan *peer.Peer)}
	for _, peer := range peerList {
		go peer.Connect([]byte(m.InfoHash), []byte(tracker.PeerID+util.SessionID(12)), t.Activate, t.Deactivate) // might want the same id for each peer
	}

	return t, nil
}

func (t *Torrent) Start() {
	// activate watcher
	go func() {
		for p := range t.Activate {
			t.Lock.Lock()
			t.ActivePeers[p.ID] = p
			log.Printf("%s is active!!", p.ID)
			t.Lock.Unlock()
		}
	}()

	// deactivate watcher
	go func() {
		for p := range t.Deactivate {
			t.Lock.Lock()
			t.ActivePeers[p.ID] = nil
			log.Printf("%s is kill!!", p.ID)
			t.Lock.Unlock()
		}
	}()

	// daemon for downloading chunks
	go func() {}()
}
