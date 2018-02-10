package shared

import "log"

type MinerRPCServer struct {
	Miner *MinerStruct
}

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	log.Println("stoped")
	m.Miner.MiningStopSig <- block
	return nil
}

// func (m *MinerRPCServer) ReceiveMinerHeartBeat(s string, alive *bool) error {
//
// }
