package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mbags/gtc/pkg/metainfo"
	"github.com/mbags/gtc/pkg/tracker"
	"github.com/mbags/gtc/pkg/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gtc <torrent>")
		return
	}
	m, err := metainfo.NewFromFilename(os.Args[1])
	if err != nil {
		log.Fatalf("Couldnt create metainfo for %v", os.Args[1])
	}
	// pretty print the parsed .torrent
	fmt.Println(m)
	// fetch & print initial peers
	peerList, err := tracker.FindPeers(m)
    if err != nil {
        log.Fatalf("Couldn't get the peers list")
        return
    }

	for _, peer := range peerList {
		go peer.Connect([]byte(m.InfoHash), []byte(tracker.PeerID+util.SessionID(12)))
	}

	select {}
}
