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
	"net/rpc"

	shared "./shared"
	am "./artminerlib"
)

//var globalInkMinerPairKey ecdsa.PrivateKey
var thisInkMiner *shared.MinerStruct
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
	//globalInkMinerPairKey = inkMinerStruct.PairKey
	fmt.Println("Miner Key: ", inkMinerStruct.PairKey.X)
  thisInkMiner =&inkMinerStruct
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

	// While the heart is beating, keep fetching for neighbours
	inkMinerStruct.CheckForNeighbour()

	// After going over the minimum neighbours value, start doing no-op
	OP := shared.Operation{Command: "no-op"}
	// i := 1
	// for {
	// fmt.Println("=============================", i)
	inkMinerStruct.Mine(OP)
	// i++
	// }

  // Listen for Art noded that want to connect to it
	fmt.Println("Going to Listen to Art Nodes: ")
	listenArtConn, err := net.Listen("tcp", "127.0.0.1:") // listening on wtv port
	CheckError(err)
	fmt.Println("Port Miner is lisening on ",  listenArtConn.Addr())

	// check that the art node has the correct public/private key pair
	initArt := new(KeyCheck)
	rpc.Register(initArt)
	cs := new(CanvasSet)
	rpc.Register(cs)
	rpc.Accept(listenArtConn)

	return
}

func initializeMiner(servAddr string, minerAddr string) shared.MinerStruct {
	minerKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	killSig := make(chan *shared.Block)
	NotEnoughNeighbourSig := make(chan bool)
	LeafMap := make(map[string]*shared.Block)
	minerNeighbourMap := make(map[string]shared.MinerStruct)
	return shared.MinerStruct{ServerAddr: servAddr,
		MinerAddr:             minerAddr,
		PairKey:               *minerKey,
		MiningStopSig:         killSig,
		NotEnoughNeighbourSig: NotEnoughNeighbourSig,
		LeafNodesMap:          LeafMap,
		FoundHash:             false,
		Neighbours:            minerNeighbourMap,
	}
}

// TODO
//type ArtNodeOpReq int
type KeyCheck int
type HeartBeat int
type CanvasSet int

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
func (l *CanvasSet) GetCanvasSettingsFromMiner(s string, ics *am.InitialCanvasSetting) error {
	fmt.Println("request for CanvasSettings")
	ics.Cs=am.CanvasSettings(thisInkMiner.Settings.CanvasSettings)
	ics.ListOfOps = thisInkMiner.ListOfOps
	fmt.Println("GetCanvasSettingsFromMiner() ", *ics)
	return nil
	}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
