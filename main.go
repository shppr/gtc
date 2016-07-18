package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mbags/gtc/pkg/metainfo"
)

func main() {
	m, err := metainfo.NewFromFilename(os.Args[1])
	if err != nil {
		log.Fatalf("Couldnt created metainfo for %v", os.Args[1])
	}
	fmt.Println(m)
}
