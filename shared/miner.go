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
func (m MinerStruct) Flood(visited *[]MinerStruct) {
	// TODO maybe an rpc call here, to stop other miner from mining
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	validNeighbours := make([]MinerStruct, 0)
	for _, m := range m.Neighbours {
		if filter(m, visited) {
			validNeighbours = append(validNeighbours, m)
		}
	}
	if len(validNeighbours) == 0 {
		return
	}
	for _, v := range validNeighbours {
		*visited = append(*visited, v)
	}
	for _, n := range validNeighbours {
		n.Flood(visited)
	}
	return
}

func filter(m MinerStruct, visited *[]MinerStruct) bool {
	for _, s := range *visited {
		if s.PublicKey == m.PublicKey {
			return false
		}
	}
	return true
}
