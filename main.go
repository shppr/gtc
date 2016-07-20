package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mbags/gtc/pkg/metainfo"
	"github.com/mbags/gtc/pkg/tracker"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage : gtc <torrent>")
		return
	}
	m, err := metainfo.NewFromFilename(os.Args[1])
	if err != nil {
		log.Fatalf("Couldnt created metainfo for %v", os.Args[1])
	}
	fmt.Println(m)
	// fetch initial peers
	peerList := tracker.FindPeers(m)
	for _, peer := range peerList {
		fmt.Printf("Peer %v\n", peer)
	}
}
