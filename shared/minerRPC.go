package shared

import "log"

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

	m.Miner.FoundHash = false
	log.Println("Sent channel info")
	return nil
}

// func (m *MinerRPCServer) ReceiveMinerHeartBeat(s string, alive *bool) error {
//
// }
