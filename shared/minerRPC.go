package shared

import "fmt"

type MinerRPCServer int

func (m *MinerRPCServer) StopMining(s string, alive *bool) error {
	fmt.Println("here")
	return nil
}
