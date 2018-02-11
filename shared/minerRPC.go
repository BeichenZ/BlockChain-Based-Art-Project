package shared

import (
	"fmt"
	"log"
	"time"
)

type MinerRPCServer struct {
	Miner *MinerStruct
}

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	log.Println("stopped")
	if !m.Miner.FoundHash {
		log.Print("I didn't find the block, so I have recieved the block from another miner")
		m.Miner.MiningStopSig <- block
	} else {
		log.Print("I have found the hash, but so did at least one other miner")
	}
	log.Println("Sent channel info")
	return nil
}

func (m *MinerRPCServer) ReceiveMinerHeartBeat(minerNeighbourAddr string, alive *bool) error {
	allNeighbour.Lock()
	defer allNeighbour.Unlock()
	log.Println("Heartbeat received from ", minerNeighbourAddr)
	fmt.Println("_____________________________________________________________________________________")
	fmt.Printf("%+v", allNeighbour.all)
	if _, ok := allNeighbour.all[minerNeighbourAddr]; ok {
		log.Println("The miner is in the map")
		allNeighbour.all[minerNeighbourAddr].RecentHeartbeat = time.Now().UnixNano()
		fmt.Println("Heartbeat RECEIVED ", allNeighbour.all[minerNeighbourAddr].RecentHeartbeat  )
	} else {
		log.Println("Nothing in the map")
	}
	return nil
}

func (m *MinerRPCServer) MinerRegister(MinerNeighbourPayload *string, alive *bool) error {
	if len(allNeighbour.all) > int(m.Miner.Settings.MinNumMinerConnections) {
		m.Miner.NotEnoughNeighbourSig <- false
	}

	if _, ok := allNeighbour.all[(*MinerNeighbourPayload)]; ok {
		log.Println("The Miner is already here")
	} else {

		allNeighbour.all[(*MinerNeighbourPayload)] = &MinerStruct{
			MinerAddr: (*MinerNeighbourPayload),
			// MinerConnection: &MinerNeighbourPayload.client,
			RecentHeartbeat: time.Now().UnixNano(),
		}
		log.Println("Registration time is: ", time.Now().UnixNano())
		log.Println("Successfully recorded this neighbouring miner", (*MinerNeighbourPayload))
		go monitor(*MinerNeighbourPayload, *m.Miner, 1000*time.Millisecond)

	}
	return nil
}
