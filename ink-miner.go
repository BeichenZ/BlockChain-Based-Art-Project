package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"os"

	shared "./shared"
	//am "./artminerlib"
)

var globalInkMinerPairKey ecdsa.PrivateKey

func main() {
	// Register necessary struct for server communications
	servAddr := "127.0.0.1:12345"
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&net.TCPAddr{})

	// Construct minerAddr from flag provided in the terminal
	minerPort := flag.String("p", "", "RPC server ip:port")
	flag.Parse()
	minerAddr := "127.0.0.1:" + *minerPort

	// initialize miner given the server address and its own miner address
	inkMinerStruct := initializeMiner(servAddr, minerAddr)
	globalInkMinerPairKey = inkMinerStruct.PairKey
	//fmt.Println("Miner Key: ", inkMinerStruct.PairKey.X)

	// RPC - Register this miner to the server
	minerSettings, error := inkMinerStruct.Register(servAddr, inkMinerStruct.PairKey.PublicKey)
	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	// setting returned from the server
	inkMinerStruct.Settings = minerSettings

	//start heartbeat to the server
	// heartBeatChannel := make(chan int)
	go inkMinerStruct.HeartBeat()
	// <-heartBeatChannel

	inkMinerStruct.CheckForNeighbour()

	// While the heart is beating, keep fetching for neighbours

	// After going over the minimum neighbours value, start doing no-op

	OP := shared.Operation{Command: "no-op"}
	inkMinerStruct.Mine(OP)

	return
}

func initializeMiner(servAddr string, minerAddr string) shared.MinerStruct {
	minerKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	killSig := make(chan *shared.Block)
	NotEnoughNeighbourSig := make(chan bool)
	LeafMap := make(map[string]*shared.Block)
	return shared.MinerStruct{ServerAddr: servAddr,
		MinerAddr:             minerAddr,
		PairKey:               *minerKey,
		MiningStopSig:         killSig,
		NotEnoughNeighbourSig: NotEnoughNeighbourSig,
		LeafNodesMap:          LeafMap,
		FoundHash:             false,
	}
}

// TODO
//type ArtNodeOpReq int
type KeyCheck int

// func (l *ArtNodeOpReq) doArtNodeOp(o am.Operation, reply *int) error {
// 	// TODO
// 	return nil
// }

// func (l *KeyCheck) ArtNodeKeyCheck(privKey ecdsa.PrivateKey, reply *bool) error {
// 	*reply = (privKey == globalInkMinerPairKey )
// 	fmt.Println("ArtNodeKeyCheck(): Art node connecting with me")
// 	return nil
// }
func (l *KeyCheck) ArtNodeKeyCheck(privKey string, reply *bool) error {
	*reply = true
	fmt.Println("ArtNodeKeyCheck(): Art node connecting with me")
	return nil
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
