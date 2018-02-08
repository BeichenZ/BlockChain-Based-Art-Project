package shared

type BlockApi interface {
	GetStringBlock() string
}

type Block struct {
	CurrentHash  string
	PreviousHash string
	OPS          []Operation
}

func (b Block) GetStringOperations() string {
	listOfOpeartionsString := b.CurrentHash + " \n "

	for _, operation := range b.OPS {
		operationString := operation.Command + "," + operation.Shapetype + " by " + operation.UserSignature + " \n "

		listOfOpeartionsString += operationString
	}

	return listOfOpeartionsString

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
