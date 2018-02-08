package main

import (
	"flag"
	"fmt"
	"encoding/gob"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net"
	"net/rpc"
	"os"

	shared "./shared"
	//am "./artminerlib"
)
var globalInkMinerPairKey ecdsa.PrivateKey
func main() {
	servAddr := "127.0.0.1:12345"
	minerPort := flag.String("p", "", "RPC server ip:port")
	flag.Parse()
	minerAddr := "127.0.0.1:" + *minerPort
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&net.TCPAddr{})

	///
	inkMinerStruct := initializeMiner(servAddr, minerAddr)
	globalInkMinerPairKey = inkMinerStruct.PairKey
	fmt.Println("Miner Key: ", inkMinerStruct.PairKey.X)

	// TODO register a miner node here, get back Neighbours info and threshold
	minerSettings, error := inkMinerStruct.Register(servAddr, inkMinerStruct.PairKey.PublicKey)

	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	inkMinerStruct.Settings = minerSettings

	minerServer := &shared.MinerRPCStruct{inkMinerStruct}

	conn, error := net.Listen("tcp", "127.0.0.1:0")

	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	rpc.Register(minerServer)
	go rpc.Accept(conn)

	// TODO start heartbeat to the server
	// heartBeatChannel := make(chan int)
	go inkMinerStruct.HeartBeat()
	// <-heartBeatChannel

	inkMinerStruct.CheckForNeighbour()
	OP := shared.Operation{Command: "no-op"}
	inkMinerStruct.Mine(OP)

	// Listen for Art noded that want to connect to it
	fmt.Println("Going to Listen to Art Nodes: ")
	listenArtConn, err := net.Listen("tcp", "127.0.0.1:") // listening on wtv port
	CheckError(err)
	fmt.Println("Port Miner is lisening on ",  listenArtConn.Addr())

	// check that the art node has the correct public/private key pair
	initArt := new(KeyCheck)
	rpc.Register(initArt)
	rpc.Accept(listenArtConn)

	// artNodeOp := new(ArtNodeOpReq)
	// rpc.Register(artNodeOp)
	// rpc.Accept(listenArtConn)


	// Command        string
	// UserSignature  string
	// AmountOfInk    int
	// Shapetype      string
	// ShapeSvgString string
	// Fill           string
	// Stroke         string

	return
}

//func startMiningForNoOp(miner shared.MinerStruct) {
//
//	noOperation := shared.Operation{"noOp", miner.PublicKey., 1, "", "", "" , ""}
//
//	for {
//		miner.Mine(noOperation)
//	}
//}

func initializeMiner(servAddr string, minerAddr string) shared.MinerStruct {
	fmt.Println("initializeMiner() ", rand.Reader)

	minerKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	return shared.MinerStruct{ServerAddr: servAddr, MinerAddr: minerAddr, PairKey: *minerKey}
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
