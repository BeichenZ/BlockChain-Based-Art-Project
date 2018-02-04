package shared

type Miner interface {
	Register(address string, publicKey string) (string, error)

	GetNodes(publicKey string) ([]string, error)

	HeartBeat(publicKey string) error

	Mine(currentBlock Block, newOperation Operation) (string, error)

	Flood(visited *[]MinerStruct) error
}

type MinerStruct struct {
	ServerAddr string
	PublicKey  string
	PrivKey    string
	Threshold  int
	Neighbours []MinerStruct
	ArtNodes   []string
}

func (m MinerStruct) Mine(currentBlock Block, newOperation Operation) (string, error) {
	return "", nil
}

func (m MinerStruct) HeartBeat(publicKey string) error {
	return nil
}

// Bare minimum flooding protocol, Miner will disseminate notification through the network
func (m MinerStruct) Flood(visited *[]MinerStruct) error {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	// TODO maybe an rpc call here, to stop other miner from mining
	nes := make([]MinerStruct, 0)
	for _, m := range m.Neighbours {
		if filter(m, visited) {
			nes = append(nes, m)
		}
	}
	*visited = append(*visited, m)
	for _, n := range nes {
		n.Flood(visited)
	}
	return nil
}

func filter(m MinerStruct, visited *[]MinerStruct) bool {
	for _, s := range *visited {
		if s.PublicKey == m.PublicKey {
			return false
		}
	}
	return true
}
