package shared

import "fmt"

type MinerRPCServer struct {
	Miner *MinerStruct
}

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	fmt.Println("stoped")
	m.Miner.MiningStopSig <- block
	return nil
}
