package main

import (
	"fmt"
	"os"

	shared "./shared"
)

func main() {
	servAddr := os.Args[1]
	publicKey := os.Args[2]
	privKey := os.Args[3]

	inkMinerStruct := initializeMiner(servAddr, publicKey, privKey)
	fmt.Println(inkMinerStruct)
	// TODO register a miner node here, get back Neighbours info and threshold

	// TODO start heartbeat to the server

	if len(inkMinerStruct.Neighbours) > inkMinerStruct.Threshold {
		// TODO start Mining for noop
	}
	return
}

func initializeMiner(servAddr string, publicKey string, privKey string) shared.MinerStruct {
	return shared.MinerStruct{ServerAddr: servAddr, PublicKey: publicKey, PrivKey: privKey}
}
