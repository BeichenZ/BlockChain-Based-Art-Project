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

	// RPC - Register this miner to the server
	minerSettings, error := inkMinerStruct.Register(servAddr, inkMinerStruct.PairKey.PublicKey)
	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	// setting returned from the server
	inkMinerStruct.Settings = minerSettings

	// RPC - Start rpc server on this ink miner
	minerServer := &shared.MinerRPCServer{Miner: &inkMinerStruct}
	rpc.Register(minerServer)
	conn, error := net.Listen("tcp", minerAddr)

	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	go rpc.Accept(conn)

	//start heartbeat to the server
	// heartBeatChannel := make(chan int)
	go inkMinerStruct.HeartBeat()
	// <-heartBeatChannel

	// While the heart is beating, keep fetching for neighbours
	inkMinerStruct.CheckForNeighbour()

	// After going over the minimum neighbours value, start doing no-op
	OP := shared.Operation{Command: "no-op"}
	inkMinerStruct.Mine(OP)

	return
}

func initializeMiner(servAddr string, minerAddr string) shared.MinerStruct {
	minerKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	killSig := make(chan bool)
	return shared.MinerStruct{ServerAddr: servAddr, MinerAddr: minerAddr, PairKey: *minerKey, MiningStopSig: killSig}
}
