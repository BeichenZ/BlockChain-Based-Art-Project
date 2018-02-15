package shared

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
	"crypto/ecdsa"
	"crypto/rand"
)

type MinerRPCServer struct {
	Miner *MinerStruct
}

type BadBlockError string

func (e BadBlockError) Error() string {
	return fmt.Sprintf("BAD BLOCK")
}

func (m *MinerRPCServer) SendChain(s string, blockChain *BlockPayloadStruct) error {
	log.Println(s)
	thisMinerBlockChain := CopyBlockChainPayload(m.Miner.BlockChain)
	*blockChain = thisMinerBlockChain
	log.Println(thisMinerBlockChain)
	return nil
}

// func (m *MinerRPCServer) SendLeafNodesMap(s string, leafNodes *map[string]*Block) error {
// 	newMap := make(map[string]*Block)
// 	for k, v := range m.Miner.LeafNodesMap {
// 		newMap[k] = CopyBlockChain(v)
// 	}
// 	*leafNodes = newMap
// 	return nil
// }

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	log.Println("stopped")
	if !m.Miner.FoundHash {
		log.Print("Someone send me a block, its hash is: ", block.CurrentHash)
		if block.Validate() {

			go func() {
				m.Miner.MiningStopSig <- block
			}()
		} else {
			return BadBlockError("BAD BLOCK")
		}
	} else {
		log.Print("I have found the hash, but so did at least one other miner")
	}
	log.Println("Sent channel info")
	return nil
}

func (m *MinerRPCServer) ReceivedOperation(operation *Operation, alive *bool) error {

	m.Miner.RecievedOpSig <- *operation

	return nil
}

func (m *MinerRPCServer) ReceiveMinerHeartBeat(minerNeighbourAddr string, alive *bool) error {
	allNeighbour.Lock()
	defer allNeighbour.Unlock()
	// log.Println("Heartbeat received from ", minerNeighbourAddr)
	// fmt.Println("_____________________________________________________________________________________")
	// fmt.Printf("%+v", allNeighbour.all)
	if _, ok := allNeighbour.all[minerNeighbourAddr]; ok {
		// log.Println("The miner is in the map")
		allNeighbour.all[minerNeighbourAddr].RecentHeartbeat = time.Now().UnixNano()
		// fmt.Println("Heartbeat RECEIVED ", allNeighbour.all[minerNeighbourAddr].RecentHeartbeat)
	} else {
		// log.Println("Nothing in the map")
	}
	return nil
}

func (m *MinerRPCServer) ActivateHeartBeat(SenderAddr string, alive *bool) error {
	go m.Miner.minerSendHeartBeat(SenderAddr)
	return nil
}

func (m *MinerRPCServer) MinerRegister(MinerNeighbourPayload *string, thisMinerChainLength *int) error {
	fmt.Println("------------------------------------I got here--------------------------------------")
	fmt.Println(MinerNeighbourPayload)
	// if len(allNeighbour.all) > int(m.Miner.Settings.MinNumMinerConnections) {
	// 	m.Miner.NotEnoughNeighbourSig <- false
	// }

	if _, ok := allNeighbour.all[(*MinerNeighbourPayload)]; ok {
		log.Println("The Miner is already here")
	} else {
		alive := false
		number := 0
		allNeighbour.all[(*MinerNeighbourPayload)] = &MinerStruct{
			MinerAddr: (*MinerNeighbourPayload),
			// MinerConnection: &MinerNeighbourPayload.client,
			RecentHeartbeat: time.Now().UnixNano(),
		}
		length := m.Miner.FindLongestChainLength()
		*thisMinerChainLength = length
		log.Println("Registration time is: ", time.Now().UnixNano())
		log.Println("Successfully recorded this neighbouring miner", (*MinerNeighbourPayload))
		client, error := rpc.Dial("tcp", *MinerNeighbourPayload)
		if error != nil {
			fmt.Println(error)
		}
		err := client.Call("MinerRPCServer.MinerRegister", m.Miner.MinerAddr, &number)
		if err != nil {
			fmt.Println(err)
		}
		go monitor(*MinerNeighbourPayload, *m.Miner, 1000*time.Millisecond)

		client.Call("MinerRPCServer.ActivateHeartBeat", m.Miner.MinerAddr, &alive)

	}
	return nil
}

type KeyCheck int
type CanvasSet struct {
	Miner MinerStruct
}
type ArtNodeOpReg struct {
	Miner MinerStruct
}

func (l *ArtNodeOpReg) DoArtNodeOp(op *Operation, reply *bool) error {

 	// check errors
 	// Insuffcient errors
 	// shape overlap
 	fmt.Println(op.Command)

 	// Sign the Operation
	r, s, err := ecdsa.Sign(rand.Reader, &l.Miner.PairKey, []byte(op.Command))

	if err != nil {
		fmt.Println(err)
	}

	op.Issuer = &l.Miner.PairKey
	op.IssuerR = r
	op.IssuerS = s
	go func (){
		l.Miner.RecievedArtNodeSig <- *op
	}()

	 blockCounter.Lock()
	 blockCounter.counter = 0
	 blockCounter.Unlock()

	// TODO  What if multiple art node tries to make operation

	 validNum := op.ValidFBlkNum
	// ** Maybe set timeout if it takes too long
	// while l.Miner.chain is not validateNum more than before don't return anything, only return when its not


	 for  {
		 //fmt.Println("DoArtNodeOp: IN loop")
		 blockCounter.Lock()
		 if blockCounter.counter == validNum {
			*reply = true
			blockCounter.Unlock()
			break
		}
		 blockCounter.Unlock()
	}
	 fmt.Println("DoArtNodeOp() validateNum condition satisfied")
 	return nil
 }

func (l *ArtNodeOpReg) ArtnodeInkRequest(s string, remainInk *uint32) error {
	*remainInk= l.Miner.MinerInk
	return nil
}

func (l *ArtNodeOpReg) ArtnodeGenBlkRequest(s string, genBlkHash *string) error {
	*genBlkHash=l.Miner.BlockChain.CurrentHash
	fmt.Println("ArtnodeGenBlkRequest() ", genBlkHash)
	return nil
}

func (l *ArtNodeOpReg) ArtnodeBlkChildRequest(s string, blkCh *[]string) error {
	
	return nil
}

func (l *KeyCheck) ArtNodeKeyCheck(privKey *string, reply *bool) error {
	*reply = true
	fmt.Println("ArtNodeKeyCheck(): Art node connecting with me")

	return nil
}
func (l *CanvasSet) GetCanvasSettingsFromMiner(s string, ics *InitialCanvasSetting) error {
	fmt.Println("request for CanvasSettings")
	ics.Cs = CanvasSettings(l.Miner.Settings.CanvasSettings)
	ics.ListOfOps_str = l.Miner.ListOfOps_str
	fmt.Println("GetCanvasSettingsFromMiner() ", *ics)
	return nil
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}