package shared

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
	am "../artminerlib"
)

type MinerRPCServer struct {
	Miner *MinerStruct
}

type BadBlockError string

func (e BadBlockError) Error() string {
	return fmt.Sprintf("BAD BLOCK")
}

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	log.Println("stopped")
	if !m.Miner.FoundHash {
		log.Print("I didn't find the block, so I have recieved the block from another miner")
		if block.Validate() {
			m.Miner.MiningStopSig <- block
		} else {
			return BadBlockError("BAD BLOCK")
		}
	} else {
		log.Print("I have found the hash, but so did at least one other miner")
	}
	log.Println("Sent channel info")
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

func (m *MinerRPCServer) MinerRegister(MinerNeighbourPayload *string, alive *bool) error {
	fmt.Println("------------------------------------I got here--------------------------------------")
	fmt.Println(MinerNeighbourPayload)
	// if len(allNeighbour.all) > int(m.Miner.Settings.MinNumMinerConnections) {
	// 	m.Miner.NotEnoughNeighbourSig <- false
	// }

	if _, ok := allNeighbour.all[(*MinerNeighbourPayload)]; ok {
		log.Println("The Miner is already here")
	} else {
		alive := false
		allNeighbour.all[(*MinerNeighbourPayload)] = &MinerStruct{
			MinerAddr: (*MinerNeighbourPayload),
			// MinerConnection: &MinerNeighbourPayload.client,
			RecentHeartbeat: time.Now().UnixNano(),
		}
		log.Println("Registration time is: ", time.Now().UnixNano())
		log.Println("Successfully recorded this neighbouring miner", (*MinerNeighbourPayload))
		client, error := rpc.Dial("tcp", *MinerNeighbourPayload)
		if error != nil {
			fmt.Println(error)
		}
		err := client.Call("MinerRPCServer.MinerRegister", m.Miner.MinerAddr, &alive)
		if err != nil {
			fmt.Println(err)
		}
		go monitor(*MinerNeighbourPayload, *m.Miner, 1000*time.Millisecond)

		client.Call("MinerRPCServer.ActivateHeartBeat", m.Miner.MinerAddr, &alive)

	}
	return nil
}



// TODO
//type ArtNodeOpReq int
type KeyCheck int
type CanvasSet struct {
	Miner MinerStruct
}
type ArtNodeOpReg struct {
	Miner MinerStruct
}

 func (l *ArtNodeOpReg) DoArtNodeOp(s *string , reply *bool) error {

 	fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	go func (){
		l.Miner.RecievedArtNodeSig <- true
	}()

 	// TODO
 	// check errors
 	// Insuffcient errors
 	// shape overlap
 	return nil
 }

// gil
// }
func (l *KeyCheck) ArtNodeKeyCheck(privKey string, reply *bool) error {
	*reply = true
	fmt.Println("ArtNodeKeyCheck(): Art node connecting with me")
	return nil
}
func (l *CanvasSet) GetCanvasSettingsFromMiner(s string, ics *am.InitialCanvasSetting) error {
	fmt.Println("request for CanvasSettings")
	ics.Cs = am.CanvasSettings(l.Miner.Settings.CanvasSettings)
	ics.ListOfOps_str = l.Miner.ListOfOps_str
	fmt.Println("GetCanvasSettingsFromMiner() ", *ics)
	return nil
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
