package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"./blockartlib"
	shared "./shared"
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
	thisInkMiner = &inkMinerStruct
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
	mux := http.NewServeMux()

	mux.HandleFunc("/getshapes", echoHandler(inkMinerStruct))
	mux.HandleFunc("/addshape", addshape(inkMinerStruct))

	http.ListenAndServe(":5000", mux)
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

func addshape(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		// TODO RPC call here
		fmt.Println(string(body))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	//Marshal or convert user object back to json and write to response
	svgPayload := SVGPayload{blockartlib.GetListOfOps()}
	s, err := json.Marshal(svgPayload)
	if err != nil {
		panic(err)
	}

	//Set Content-Type header so that clients will know how to read response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	//Write json response back to response
	w.Write(s)
}
