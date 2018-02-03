package main


type Miner interface {

	Register(address string, publicKey string) (string, error)

	GetNodes(publicKey string) ([]string, error)

	HeartBeat(publicKey string) (error)

	Mine(currentBlock Block, newOperation Operation) (string, error)

}



