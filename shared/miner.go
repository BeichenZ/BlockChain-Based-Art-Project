package shared

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
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

type MinerStruct struct {
	ServerAddr            string
	MinerAddr             string
	PairKey               ecdsa.PrivateKey
	Threshold             int
	Neighbours            map[string]MinerStruct
	ArtNodes              []string
	BlockChain            *Block
	ServerConnection      *rpc.Client
	MinerConnection       *rpc.Client
	Settings              MinerNetSettings
	MiningStopSig         chan *Block
	NotEnoughNeighbourSig chan bool
	LeafNodesMap          map[string]*Block
	FoundHash             bool
	RecentHeartbeat       int64
	ListOfOps []string
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
	for _, v := range m.LeafNodesMap {
		if v.DistanceToGenesis > localMax {
			fmt.Println("Finding the leading block: The hash is" + v.CurrentHash)
			localMax = v.DistanceToGenesis
			maxBlock = v
		}
	}

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
	m.LeafNodesMap[genesisBlock.CurrentHash] = &genesisBlock
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

func (m *MinerStruct) Mine(newOperation Operation) (string, error) {
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
			nonce := leadingBlock.GetNonce()
			nonce += newOperation.Command + "," + newOperation.Shapetype + " by " + newOperation.UserSignature + " \n "

			newBlock := doProofOfWork(m, nonce, 4, 100, newOperation, leadingBlock)
			leadingBlock.Children = append(leadingBlock.Children, newBlock)
			// TODO maybe validate block here
			printBlock(m.BlockChain)
			fmt.Println("\n")
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
func (m MinerStruct) Flood(newBlock *Block, visited *[]MinerStruct) {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO what happense if node A calls flood, and before it can reach node B, node B calls flood?
	validNeighbours := make([]MinerStruct, 0)
	fmt.Println("Flooding is called.......................................................")
	for _, v := range m.Neighbours {
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
		alive := false
		fmt.Println("visiting miner: ", n.MinerAddr)
		passingBlock := copyBlock(newBlock)
		err := n.MinerConnection.Call("MinerRPCServer.StopMining", passingBlock, &alive)
		if err != nil {
			fmt.Println(err)
		}
		n.Flood(newBlock, visited)
	}
	return
}

func (m *MinerStruct) produceBlock(currentHash string, newOP Operation, leadingBlock *Block) *Block {
	// visitedMiners := make([]MinerStruct, 0)
	visitedMiners := []MinerStruct{*m}
	/// Find the leading block
	LocalOPs := []Operation{newOP}
	fmt.Println("Creating a new block with the new hash")
	producedBlock := &Block{CurrentHash: currentHash,
		PreviousHash:      leadingBlock.CurrentHash,
		LocalOPs:          LocalOPs,
		Children:          make([]*Block, 0),
		DistanceToGenesis: leadingBlock.DistanceToGenesis + 1}
	m.Flood(producedBlock, &visitedMiners)

	delete(m.LeafNodesMap, leadingBlock.CurrentHash)
	fmt.Println("Need to let the other miners about this block")
	m.FoundHash = false
	m.LeafNodesMap[producedBlock.CurrentHash] = producedBlock
	return producedBlock
}

func (m *MinerStruct) CheckForNeighbour() {
	listofNeighbourIP := make([]net.Addr, 0)
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
			fmt.Println("Fucked; can't connect")
			fmt.Println(error)
			log.Fatal(error)
			os.Exit(0)
		}

		if _, exists := m.Neighbours[netIP.String()]; exists {
			fmt.Println("neighbour exist and alive")
		} else {
			m.Neighbours[netIP.String()] = MinerStruct{
				MinerAddr:       netIP.String(),
				MinerConnection: client,
				RecentHeartbeat: time.Now().UnixNano(),
			}
			go monitor(netIP.String(), *m, 500)
		}

	}
}
