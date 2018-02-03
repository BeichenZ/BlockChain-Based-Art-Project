package shared

type Miner interface {
	Register(address string, publicKey string) (string, error)

	GetNodes(publicKey string) ([]string, error)

	HeartBeat(publicKey string) error

	Mine(currentBlock Block, newOperation Operation) (string, error)

	Flood(senders []MinerStruct) error
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
func (m MinerStruct) Flood(senders []MinerStruct) error {
	// TODO construct a list of MinerStruct excluding the senders to avoid infinite loop
	nes := make([]MinerStruct, 0)
	for _, m := range m.Neighbours {
		if filter() {
			nes = append(nes, m)
		}
	}
	senders = append(senders, m)
	for _, n := range nes {

		n.Flood(senders)
	}
	return nil
}

func filter() bool {
	return false
}
