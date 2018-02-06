package shared

type MinerRPCStruct struct {
	Miner MinerStruct
}

func (m *MinerRPCStruct) StopMining(miner MinerStruct, r *MinerStruct) error {
	// TODO called inside the flooding protocol, stop this miner from Mining
	return nil
}
