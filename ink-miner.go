package main

import (
	"crypto/md5"
	"encoding/hex"
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
		// TODO start Mining for no-op
		// TODO break the mining loop once a secret is found
		// TODO flood the network
	}

	return
}

func initializeMiner(servAddr string, publicKey string, privKey string) shared.MinerStruct {
	return shared.MinerStruct{ServerAddr: servAddr, PublicKey: publicKey, PrivKey: privKey}
}

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}
