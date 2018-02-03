package shared

type Miner interface {
	Register(address string, publicKey string) (string, error)

	GetNodes(publicKey string) ([]string, error)

	HeartBeat(publicKey string) error

	Mine(currentBlock Block, newOperation Operation) (string, error)
}

type MinerStruct struct {
	ServerAddr string
	PublicKey  string
	PrivKey    string
	Threshold  int
	Neighbours []string
	ArtNodes   []string
}
