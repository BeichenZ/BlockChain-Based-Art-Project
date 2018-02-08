package shared

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

var (
	globalStop = true
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
	ServerAddr    string
	MinerAddr     string
	PairKey       ecdsa.PrivateKey
	Threshold     int
	Neighbours    []MinerStruct
	ArtNodes      []string
	BlockChain    *Block
	Client        *rpc.Client
	Settings      MinerNetSettings
	MiningStopSig chan bool
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

func (m *MinerStruct) Register(address string, publicKey ecdsa.PublicKey) (MinerNetSettings, error) {
	// fmt.Println("public key", publicKey)
	client, error := rpc.Dial("tcp", address)
	minerSettings := &MinerNetSettings{}
	if error != nil {
		return *minerSettings, error
	}

	m.Client = client

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
	return *minerSettings, err
}

func (m MinerStruct) HeartBeat() error {
	alive := false

	for {
		error := m.Client.Call("RServer.HeartBeat", m.PairKey.PublicKey, &alive)
		if error != nil {
			fmt.Println(error)
		}
		time.Sleep(time.Millisecond * time.Duration(800))
	}
}

func (m *MinerStruct) Mine(newOperation Operation) (string, error) {
	// currentBlock := m.BlockChain[len(m.BlockChain)-1]
	// listOfOperation := currentBlock.GetStringOperations()
	listOfOperation := ""
	listOfOperation += newOperation.Command + "," + newOperation.Shapetype + " by " + newOperation.UserSignature + " \n "

	newHash := doProofOfWork(m, listOfOperation, 4, 100)
	fmt.Println(newHash)
	// newOperationsList := append(currentBlock.OPS, newOperation)
	//
	// newBlock := Block{newHash, currentBlock.CurrentHash, newOperationsList}
	//
	// m.BlockChain = append(m.BlockChain, newBlock)

	// update all its neighbours
	visitedMiners := make([]MinerStruct, 0)
	m.Flood(&visitedMiners)

	return "", nil
}

// Bare minimum flooding protocol, Miner will disseminate notification through the network
func (m MinerStruct) Flood(visited *[]MinerStruct) {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO what happense if node A calls flood, and before it can reach node B, node B calls flood?
	validNeighbours := make([]MinerStruct, 0)
	for _, n := range m.Neighbours {
		if filter(n, visited) {
			validNeighbours = append(validNeighbours, n)
		}
	}
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
		}

		alive := false
		fmt.Println("visiting miner: ", n.MinerAddr)
		err := client.Call("MinerRPCServer.StopMining", "stop", &alive)
		if err != nil {
			fmt.Println(err)
		}
		n.Flood(visited)
	}
	return
}

func filter(m MinerStruct, visited *[]MinerStruct) bool {
	for _, s := range *visited {
		if s.MinerAddr == m.MinerAddr {
			return false
		}
	}
	return true
}

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	fmt.Println(str)
	return str
}

func doProofOfWork(m *MinerStruct, nonce string, numberOfZeroes int, delay int) string {
	i := int64(0)

	var zeroesBuffer bytes.Buffer
	for i := int64(0); i < int64(numberOfZeroes); i++ {
		zeroesBuffer.WriteString("0")
	}
	zeroes := zeroesBuffer.String()

	for {
		select {
		case <-m.MiningStopSig:
			fmt.Println(m.MiningStopSig)
			fmt.Println("I'm DONE")
			return ""
		default:
			guessString := strconv.FormatInt(i, 10)
			if computeNonceSecretHash(nonce, guessString)[32-numberOfZeroes:] == zeroes {
				fmt.Println(guessString)
				return guessString
			}
			i++
			if m.MinerAddr[len(m.MinerAddr)-1:] == "8" {
				time.Sleep(time.Millisecond * time.Duration(delay))
			}
		}
	}
}

func (m *MinerStruct) CheckForNeighbour() {
	listofNeighbourIP := make([]net.Addr, 0)
	// var listofNeighbourIP []net.Addr
	for len(listofNeighbourIP) < int(m.Settings.MinNumMinerConnections) {
		error := m.Client.Call("RServer.GetNodes", m.PairKey.PublicKey, &listofNeighbourIP)
		if error != nil {
			fmt.Println(error)
		}
	}

	minerNeighbours := make([]MinerStruct, 0)
	for _, m := range listofNeighbourIP {
		fmt.Println(m.String())
		minerNeighbours = append(minerNeighbours, MinerStruct{MinerAddr: m.String()})
	}
	m.Neighbours = minerNeighbours
}
