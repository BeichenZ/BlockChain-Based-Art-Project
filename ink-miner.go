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
	fmt.Printf("%+v", inkMinerStruct)
	return
}

func initializeMiner(servAddr string, publicKey string, privKey string) shared.MinerStruct {
	return shared.MinerStruct{ServerAddr: servAddr, PublicKey: publicKey, PrivKey: privKey}
}
