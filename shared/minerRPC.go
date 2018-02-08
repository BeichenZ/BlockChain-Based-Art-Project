package shared

import "fmt"

type MinerRPCServer struct {
	Miner *MinerStruct
}

func (m *MinerRPCServer) StopMining(s string, alive *bool) error {
	fmt.Println(s)
	m.Miner.MiningStopSig <- true
	return nil
}
