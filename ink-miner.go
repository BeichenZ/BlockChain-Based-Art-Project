package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	//"net/http"

	"net/rpc"
	"os"

	shared "./shared"
)

//var globalInkMinerPairKey ecdsa.PrivateKey
var thisInkMiner *shared.MinerStruct

func main() {
	// Register necessary struct for server communications
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&net.TCPAddr{})

	// Construct minerAddr from flag provided in the terminal
	minerPort := flag.String("p", "", "RPC server ip:port")
	servAddr := flag.String("sa", "", "Server ip address")
	flag.Parse()
	minerAddr := "127.0.0.1:" + *minerPort

	// initialize miner given the server address and its own miner address
	inkMinerStruct := initializeMiner(*servAddr, minerAddr)
	//globalInkMinerPairKey = inkMinerStruct.PairKey
	fmt.Println("Miner Key: ", inkMinerStruct.PairKey.X)
	thisInkMiner = &inkMinerStruct
	// RPC - Register this miner to the server
	minerSettings, error := inkMinerStruct.Register(*servAddr, inkMinerStruct.PairKey.PublicKey)
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

	// Listen for Art noded that want to connect to it
	fmt.Println("Going to Listen to Art Nodes: ")
	listenArtConn, err := net.Listen("tcp", "127.0.0.1:") // listening on wtv port
	shared.CheckError(err)
	fmt.Println("Port Miner is lisening on ", listenArtConn.Addr())

	// check that the art node has the correct public/private key pair
	initArt := new(shared.KeyCheck)
	rpc.Register(initArt)
	cs := &shared.CanvasSet{inkMinerStruct}
	rpc.Register(cs)
	anr := &shared.ArtNodeOpReg{&inkMinerStruct}
	go rpc.Register(anr)
	go rpc.Accept(listenArtConn)

	// While the heart is beating, keep fetching for neighbours

	// After going over the minimum neighbours value, start doing no-op

	OP := shared.Operation{ShapeSvgString: "no-op"}
	for {
		inkMinerStruct.CheckForNeighbour()
		inkMinerStruct.StartMining(OP)
	}
	return
}

func initializeMiner(servAddr string, minerAddr string) shared.MinerStruct {
	minerKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	killSig := make(chan *shared.Block)
	NotEnoughNeighbourSig := make(chan bool)
	RecievedArtNodeSig := make(chan shared.Operation)
	RecievedOpSig := make(chan shared.Operation)

	return shared.MinerStruct{ServerAddr: servAddr,
		MinerAddr:             minerAddr,
		PairKey:               *minerKey,
		MiningStopSig:         killSig,
		NotEnoughNeighbourSig: NotEnoughNeighbourSig,
		FoundHash:             false,
		RecievedArtNodeSig:    RecievedArtNodeSig,
		RecievedOpSig:         RecievedOpSig,
	}
}
