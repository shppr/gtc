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
	t, err := torrent.NewFromFilename(os.Args[1])
	if err != nil {
		panic(err)
	}
	go t.Start()
	select {}
}
