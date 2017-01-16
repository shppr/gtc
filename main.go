package main

import (
	"fmt"
	"os"

	"github.com/mbags/gtc/pkg/torrent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gtc <torrent>")
		return
	}
	// m, err := metainfo.NewFromFilename(os.Args[1])
	// if err != nil {
	// 	log.Fatalf("Couldnt create metainfo for %v", os.Args[1])
	// }
	// // pretty print the parsed .torrent
	// fmt.Println(m)
	// // fetch & print initial peers
	// peerList, err := tracker.FindPeers(m)
	// if err != nil {
	//     log.Fatalf("Couldn't get the peers list")
	//     return
	// }

	// for _, peer := range peerList {
	// 	go peer.Connect([]byte(m.InfoHash), []byte(tracker.PeerID+util.SessionID(12)))
	// }
	t, err := torrent.NewFromFilename(os.Args[1])
	if err != nil {
		panic(err)
	}
	go t.Start()
	select {}
}
