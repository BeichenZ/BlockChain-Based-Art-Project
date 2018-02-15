package shared

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
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
	Cmd rune
	X int
	Y int
	ValCnt int
}

// One operation contains multiple movs
type SingleOp struct {
	IsClosedShape bool
	MovList       []SingleMov
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
	OPBuffer			  []Operation

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

	genesisBlock := Block{CurrentHash: minerSettings.GenesisBlockHash, Children: make([]*Block, 0)}
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
	for _, op := range buffer{
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
				fmt.Println(leadingBlock)
				// nonce := leadingBlock.GetString()
			var nonce string
			if len(m.OPBuffer) == 0 {
				//	Mine for no-op
				fmt.Println("Start doing no-op")
				nonce = initialOP.Command + pubKeyToString(m.PairKey.PublicKey) + leadingBlock.CurrentHash
			} else {
				nonce = AllOperationsCommands(m.OPBuffer) + pubKeyToString(m.PairKey.PublicKey) + leadingBlock.CurrentHash
				log.Println("Loggin out what's in the buffer")
				fmt.Println(AllOperationsCommands(m.OPBuffer))
				m.OPBuffer = make([]Operation, 0)
			}
			newBlock := doProofOfWork(m, nonce, 5, 100, initialOP, leadingBlock)
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



func (m *MinerStruct) produceBlock(currentHash string, newOP Operation, leadingBlock *Block) *Block {
	// visitedMiners := make([]MinerStruct, 0)
	visitedMiners := []*MinerStruct{m}
	/// Find the leading block
	// CurrentOPs := []Operation{newOP}
	r, s, err := ecdsa.Sign(rand.Reader, &m.PairKey, []byte(currentHash))
	if err != nil {
		os.Exit(500)
	}
	fmt.Println("Creating a new block with the new hash")
	producedBlock := &Block{CurrentHash: currentHash,
		PreviousHash: leadingBlock.CurrentHash,
		CurrentOP:    newOP,
		UserSignature: UserSignatureSturct{
			r: r,
			s: s,
		},
		Children:          make([]*Block, 0),
		DistanceToGenesis: leadingBlock.DistanceToGenesis + 1}
	m.Flood(producedBlock, &visitedMiners)
	LeafNodesMap.Lock()
	delete(LeafNodesMap.all, leadingBlock.CurrentHash)
	fmt.Println("Need to let the other miners about this block")
	m.FoundHash = false
	LeafNodesMap.all[producedBlock.CurrentHash] = producedBlock
	LeafNodesMap.Unlock()
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
	// NeighbourMap := reflect.ValueOf(allNeighbour).MapKeys()
	// NeighbourSlice := make([]net.Addr, 0)
	var listofNeighbourIP = make([]net.Addr, 0)
	for len(listofNeighbourIP) < int(m.Settings.MinNumMinerConnections) {
		error := m.ServerConnection.Call("RServer.GetNodes", m.PairKey.PublicKey, &listofNeighbourIP)
		if error != nil {
			fmt.Println(error)
		}
	}
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
		alive := false
		// payLoad := MinerHeartbeatPayload{MinerAddr: netIP.String(), client: *client}
		log.Println("NETIP IS ", netIP.String())
		client.Call("MinerRPCServer.MinerRegister", m.MinerAddr, &alive)

		for {
			if _, exists := allNeighbour.all[netIP.String()]; exists {
				fmt.Printf("The neighbour %v has registered as client", netIP.String())
				break
			}
		}
	}
}
