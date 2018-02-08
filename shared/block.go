package shared

type BlockApi interface {
	GetStringBlock() string
}

type Block struct {
	CurrentHash       string
	PreviousHash      string
	LocalOPs          []Operation
	Children          []*Block
	Parent            *Block
	DistanceToGenesis int
}

func (b Block) GetNonce() string {
	nonce := b.CurrentHash + " \n "

	for _, operation := range b.LocalOPs {
		operationString := operation.Command + "," + operation.Shapetype + " by " + operation.UserSignature + " \n "

		nonce += operationString
	}

	return nonce

}

func (b Block) UpdateDistanceToGenesis() {
	//assume only one longest chain for now
	b.DistanceToGenesis = b.Parent.DistanceToGenesis + 1
}

type Operation struct {
	Command        string
	UserSignature  string
	AmountOfInk    int
	Shapetype      string
	ShapeSvgString string
	Fill           string
	Stroke         string
}
