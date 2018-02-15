package shared

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
	"strconv"
)

type AllNeighbour struct {
	sync.RWMutex
	all map[string]*MinerStruct
}

type LeafNodes struct {
	sync.RWMutex
	all map[string]*Block
}

type GlobalBlockCreationCounter struct {
	sync.RWMutex
	counter uint8
}

var (
	allNeighbour AllNeighbour = AllNeighbour{all: make(map[string]*MinerStruct)}
	LeafNodesMap LeafNodes =  LeafNodes{all: make(map[string]*Block) }
	blockCounter = GlobalBlockCreationCounter{counter: 0}
	)


type Miner interface {
	Register(address string, publicKey ecdsa.PublicKey) (*MinerNetSettings, error)

	GetNodes(publicKey ecdsa.PublicKey) ([]string, error)

	HeartBeat(publicKey ecdsa.PublicKey, interval uint32) error

	Mine(newOperation Operation) (string, error)

	CheckforNeighbours() bool

	Flood(visited *[]MinerStruct) error

	// RPC methods of Miner
	StopMining(miner MinerStruct, r *MinerStruct) error
}

//Struct for descripting Geometry
type Point struct {
	X, Y int
}

//one move , represent like : m 100 100
type SingleMov struct {
	Cmd    rune
	X      int
	Y      int
	ValCnt int
}

// One operation contains multiple movs
type SingleOp struct {
	IsClosedShape bool
	MovList       []SingleMov
}

type BlockPayloadStruct struct {
	CurrentHash       string
	PreviousHash      string
	R                 big.Int
	S                 big.Int
	CurrentOPs        []Operation
	Children          []BlockPayloadStruct
	DistanceToGenesis int
	Nonce             int32
	SolverPublicKey   string //Make this field a string so no more seg fault
}

type MinerStruct struct {
	ServerAddr            string
	MinerAddr             string
	PairKey               ecdsa.PrivateKey
	Threshold             int
	ArtNodes              []string
	BlockChain            *Block
	ServerConnection      *rpc.Client
	MinerConnection       *rpc.Client
	Settings              MinerNetSettings
	MiningStopSig         chan *Block
	NotEnoughNeighbourSig chan bool
	FoundHash             bool
	RecentHeartbeat       int64
	ListOfOps_str         []string
	RecievedArtNodeSig    chan Operation
	RecievedOpSig         chan Operation
	OPBuffer              []Operation
}

type MinerHeartbeatPayload struct {
	client    rpc.Client
	MinerAddr string
}
type MinerInfo struct {
	Address net.Addr
	Key     ecdsa.PublicKey
}



type MinerSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`
}
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32 `json:"canvas-x-max"`
	CanvasYMax uint32 `json:"canvas-y-max"`
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	MinerSettings

	// Canvas settings
	CanvasSettings CanvasSettings `json:"canvas-settings"`
}

func copyBigInt(b *big.Int) int64 {
	return b.Int64()
}


func (m *MinerStruct) FindtheLeadingBlock() []*Block {

	var maxBlock *Block
	localMax := -1
	LeafNodesMap.Lock()
	for _, v := range LeafNodesMap.all {
		if v.DistanceToGenesis > localMax {
			fmt.Println("Finding the leading block: The hash is" + v.CurrentHash)
			localMax = v.DistanceToGenesis
			maxBlock = v
		}
	}

	LeafNodesMap.Unlock()

	thing := []*Block{maxBlock}
	return thing
}

func (m *MinerStruct) FindLongestChainLength() int {

	localMax := -1
	LeafNodesMap.Lock()
	for _, v := range LeafNodesMap.all {
		if v.DistanceToGenesis > localMax {
			fmt.Println("Finding the leading block: The hash is" + v.CurrentHash)
			localMax = v.DistanceToGenesis

		}
	}
	LeafNodesMap.Unlock()

	return localMax
}

func (m *MinerStruct) Register(address string, publicKey ecdsa.PublicKey) (MinerNetSettings, error) {
	// fmt.Println("public key", publicKey)
	///

	// RPC - Start rpc server on this ink miner
	minerServer := &MinerRPCServer{Miner: m}
	rpc.Register(minerServer)
	conn, error := net.Listen("tcp", m.MinerAddr)

	if error != nil {
		fmt.Println(error.Error())
		os.Exit(0)
	}

	go rpc.Accept(conn)

	client, error := rpc.Dial("tcp", address)
	minerSettings := &MinerNetSettings{}
	if error != nil {
		return *minerSettings, error
	}

	m.ServerConnection = client

	// RPC to server
	minerAddress, err := net.ResolveTCPAddr("tcp", m.MinerAddr)

	if err != nil {
		return *minerSettings, err
	}

	minerInfo := &MinerInfo{minerAddress, publicKey}
	err = client.Call("RServer.Register", minerInfo, minerSettings)

	if err != nil {
		return *minerSettings, err
	}
	// CurrentHash  string
	// PreviousHash string
	// // UserSignature     UserSignatureSturct
	// R                 *big.Int
	// S                 *big.Int
	// CurrentOP         Operation
	// Children          []*Block
	// DistanceToGenesis int
	// Nonce             int32
	// SolverPublicKey   *ecdsa.PublicKey
	genesisBlock := Block{
		CurrentHash:       minerSettings.GenesisBlockHash,
		PreviousHash:      "",
		R:                 &big.Int{},
		S:                 &big.Int{},
		CurrentOPs:         make([]Operation,0),
		DistanceToGenesis: 0,
		Nonce:             int32(0),
		Children:          make([]*Block, 0),
		SolverPublicKey: &ecdsa.PublicKey{
			Curve: elliptic.P384(),
			X:     &big.Int{},
			Y:     &big.Int{},
		},
	}
	m.BlockChain = &genesisBlock
	LeafNodesMap.Lock()
	LeafNodesMap.all[genesisBlock.CurrentHash] = &genesisBlock
	LeafNodesMap.Unlock()
	return *minerSettings, err
}

func (m MinerStruct) HeartBeat() error {
	alive := false

	for {
		error := m.ServerConnection.Call("RServer.HeartBeat", m.PairKey.PublicKey, &alive)
		if error != nil {
			fmt.Println(error)
		}
		time.Sleep(time.Millisecond * time.Duration(800))
	}
}

func AllOperationsCommands(buffer []Operation) string {
	retstring := ""
	for _, op := range buffer {
		retstring += op.Command
	}
	return retstring
}
func (m *MinerStruct) StartMining(initialOP Operation) (string, error) {
	// currentBlock := m.BlockChain[len(m.BlockChain)-1]
	// listOfOperation := currentBlock.GetStringOperations()
	for {
		select {
		case <-m.NotEnoughNeighbourSig:
			fmt.Println("not enough neighbour, stop minging here")
			// delete(m.LeafNodesMap, leadingBlock.CurrentHash)
			// m.LeafNodesMap[recievedBlock.CurrentHash] = recievedBlock
			return "", nil
		default:
			fmt.Println("I'm starting to mine")
			leadingBlock := m.FindtheLeadingBlock()[0]
			//
			// fmt.Println("Logging out leading block here")
			// fmt.Println(leadingBlock)

			var nonce string
			listOfOpeartion := make([]Operation,0)
			if len(m.OPBuffer) == 0 {
				//	Mine for no-op
				fmt.Println("Start doing no-op")
				nonce = leadingBlock.CurrentHash + initialOP.Command + pubKeyToString(m.PairKey.PublicKey)
				listOfOpeartion = append(listOfOpeartion, Operation{Command: "no-op"})
			} else {
				nonce = leadingBlock.CurrentHash + AllOperationsCommands(m.OPBuffer) + pubKeyToString(m.PairKey.PublicKey)
				log.Println("Loggin out what's in the buffer")
				//fmt.Println(AllOperationsCommands(m.OPBuffer))
				fmt.Println(leadingBlock.CurrentHash + AllOperationsCommands(m.OPBuffer) + pubKeyToString(m.PairKey.PublicKey))
				listOfOpeartion = m.OPBuffer
				m.OPBuffer = make([]Operation, 0)



			}
			newBlock := doProofOfWork(m, nonce, 5, 100, listOfOpeartion, leadingBlock)
			blockCounter.Lock()
			blockCounter.counter++
			blockCounter.Unlock()

				leadingBlock.Children = append(leadingBlock.Children, newBlock)
				// TODO maybe validate block here
				// printBlock(m.BlockChain)
				fmt.Println("\n")


			// time.Sleep(5000 * time.Millisecond)

			// if m.MinerAddr[len(m.MinerAddr)-1:] == "8" {
			// 	time.Sleep(time.Millisecond * time.Duration(delay))
			// }
		}
	}

	// newOperationsList := append(currentBlock.OPS, newOperation)
	//
	// newBlock := Block{newHash, currentBlock.CurrentHash, newOperationsList}
	//
	// m.BlockChain = append(m.BlockChain, newBlock)

	// update all its neighbours

	// return "", nil
}

// Bare minimum flooding protocol, Miner will disseminate notification through the network
func (m MinerStruct) Flood(newBlock *Block, visited *[]*MinerStruct) {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO what happense if node A calls flood, and before it can reach node B, node B calls flood?
	validNeighbours := make([]*MinerStruct, 0)
	fmt.Println("Flooding is called.......................................................")
	for _, v := range allNeighbour.all {
		if filter(v, visited) {
			validNeighbours = append(validNeighbours, v)
		}
	}
	fmt.Println("valid nei", len(validNeighbours))
	if len(validNeighbours) == 0 {

		return
	}

	for _, v := range validNeighbours {
		*visited = append(*visited, v)
	}
	for _, n := range validNeighbours {
		client, error := rpc.Dial("tcp", n.MinerAddr)
		if error != nil {
			fmt.Println(error)
			return
		}

		alive := false
		fmt.Println("visiting miner: ", n.MinerAddr)
		// passingBlock := copyBlock(newBlock)
		err := client.Call("MinerRPCServer.StopMining", newBlock, &alive)
		if err != nil {
			fmt.Println(err)
		}
		n.Flood(newBlock, visited)
	}
	return
}

func (m MinerStruct) FloodOperation(newOP *Operation, visited *[]*MinerStruct) {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO what happense if node A calls flood, and before it can reach node B, node B calls flood?
	validNeighbours := make([]*MinerStruct, 0)
	fmt.Println("Flooding is called.......................................................")
	for _, v := range allNeighbour.all {
		if filter(v, visited) {
			validNeighbours = append(validNeighbours, v)
		}
	}
	fmt.Println("valid nei", len(validNeighbours))
	if len(validNeighbours) == 0 {

		return
	}

	for _, v := range validNeighbours {
		*visited = append(*visited, v)
	}
	for _, n := range validNeighbours {
		client, error := rpc.Dial("tcp", n.MinerAddr)
		if error != nil {
			fmt.Println(error)
			return
		}

		alive := false
		fmt.Println("visiting miner: ", n.MinerAddr)
		// passingBlock := copyBlock(newBlock)
		err := client.Call("MinerRPCServer.ReceivedOperation", newOP, &alive)
		if err != nil {
			fmt.Println(err)
		}
		n.FloodOperation(newOP, visited)
	}
	return
}

func (m *MinerStruct) produceBlock(currentHash string, newOPs []Operation, leadingBlock *Block, nonce string) *Block {
	// visitedMiners := make([]MinerStruct, 0)
	visitedMiners := []*MinerStruct{m}
	/// Find the leading block
	// CurrentOPs := []Operation{newOP}
	r, s, err := ecdsa.Sign(rand.Reader, &m.PairKey, []byte(currentHash))
	if err != nil {
		os.Exit(500)
	}
	fmt.Println("Creating a new block with the new hash")


	sss, err := strconv.Atoi(nonce)

	if err != nil {

		fmt.Println(err)
	}

	fmt.Println("THE NONCE IS ", nonce)

	fmt.Println("THE INT NONCE IS ", sss)
	fmt.Println("THE INT32 NONCE IS ", int32(sss))

	// TODO this is not right, we're putting solver's publicKey in producedBlock
	producedBlock := &Block{CurrentHash: currentHash,
		PreviousHash: leadingBlock.CurrentHash,
		CurrentOPs:    newOPs,
		R:            r,
		S:            s,
		Children:          make([]*Block, 0),
		SolverPublicKey:   &m.PairKey.PublicKey,
		DistanceToGenesis: leadingBlock.DistanceToGenesis + 1,
		Nonce:              int32(sss) }
	m.Flood(producedBlock, &visitedMiners)
	LeafNodesMap.Lock()
	delete(LeafNodesMap.all, leadingBlock.CurrentHash)
	fmt.Println("Need to let the other miners about this block")
	m.FoundHash = false
	LeafNodesMap.all[producedBlock.CurrentHash] = producedBlock
	LeafNodesMap.Unlock()
	fmt.Println("I have found the hash, this is my public key")
	fmt.Printf("%+v", producedBlock.SolverPublicKey)
	printBlock(m.BlockChain)
	return producedBlock
}

func (m *MinerStruct) minerSendHeartBeat(minerNeighbourAddr string) error {
	alive := false
	fmt.Println(minerNeighbourAddr)
	// fmt.Println("MAKING RPC CALL TO NEIGHBOUR ", minerNeighbourAddr)
	client, _ := rpc.Dial("tcp", minerNeighbourAddr)
	for {
		fmt.Println("sending heartbeat")
		// fmt.Println(minerToMinerConnection)
		err := client.Call("MinerRPCServer.ReceiveMinerHeartBeat", m.MinerAddr, &alive)
		if err == nil {
			// fmt.Println("////////////////////////////////////////////////////////////////")
			log.Println(err)
		} else {
			return err
		}
		time.Sleep(time.Millisecond * time.Duration(400))
	}
}

func (m *MinerStruct) CheckForNeighbour() {
	var listofNeighbourIP = make([]net.Addr, 0)
	for len(listofNeighbourIP) < int(m.Settings.MinNumMinerConnections) {
		error := m.ServerConnection.Call("RServer.GetNodes", m.PairKey.PublicKey, &listofNeighbourIP)
		if error != nil {
			fmt.Println(error)
		}
	}
	localMax := -1
	var neighbourWithLongestChain string
	blockChain := BlockPayloadStruct{}
	nodesMap := make(map[string]*Block)
	for _, netIP := range listofNeighbourIP {

		fmt.Println("neighbour ip address", netIP.String())
		client, error := rpc.Dial("tcp", netIP.String())
		fmt.Println(client)

		if error != nil {
			fmt.Println(" can't connect")
			fmt.Println(error)
			log.Fatal(error)
			os.Exit(0)
		}
		neighbourBlockChainLength := 0
		// payLoad := MinerHeartbeatPayload{MinerAddr: netIP.String(), client: *client}
		log.Println("NETIP IS ", netIP.String())
		client.Call("MinerRPCServer.MinerRegister", m.MinerAddr, &neighbourBlockChainLength)
		log.Println("the neighbour's blockchain length is: ", neighbourBlockChainLength)
		for {
			if _, exists := allNeighbour.all[netIP.String()]; exists {
				fmt.Printf("The neighbour %v has registered as client", netIP.String())
				break
			}
		}

		if neighbourBlockChainLength > localMax {
			localMax = neighbourBlockChainLength
			neighbourWithLongestChain = netIP.String()
		}
	}
	// TODO get the chain from the neighbour with the longest chain
	longClient, err := rpc.Dial("tcp", neighbourWithLongestChain)
	log.Println("Connected to the longest client")
	if err != nil {
		log.Println(err)
	}
	longClient.Call("MinerRPCServer.SendChain", "give me your chain", &blockChain)
	m.BlockChain = ParseBlockChain(blockChain)
	log.Println("received block chain")
	log.Println(m.BlockChain)
	// TODO request for LeafNodesMap, bugs happen because this miner's leafnodemap is not updated
	longClient.Call("MinerRPCServer.SendLeafNodesMap", "give me your leaf", &nodesMap)
	log.Println("received leaf nodes map")
	for _, node := range nodesMap {
		log.Println(&node)
	}
	// The bug is caused by m.Blockchain and m.LeftNodesMap pointing to different address
	LeafNodesMap.Lock()
	LeafNodesMap.all[deepestBlock(m.BlockChain).CurrentHash] = deepestBlock(m.BlockChain)
	LeafNodesMap.Unlock()

}
