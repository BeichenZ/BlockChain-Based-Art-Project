package main

import (
	"flag"
	"fmt"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"net"
	"net/rpc"
	"os"

	shared "./shared"
)

func main() {
	servAddr := "127.0.0.1:12345"
	minerPort := flag.String("p", "", "RPC server ip:port")
	flag.Parse()
	minerAddr := "127.0.0.1:" + *minerPort
	gob.Register(&elliptic.CurveParams{})
	gob.Register(&net.TCPAddr{})

	///
	inkMinerStruct := initializeMiner(servAddr, minerAddr)

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

	minerKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	return shared.MinerStruct{ServerAddr: servAddr, MinerAddr: minerAddr, PairKey: *minerKey}
}
