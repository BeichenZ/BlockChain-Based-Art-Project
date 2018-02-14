package shared

import (
	"crypto/ecdsa"
)

type Operation struct {
	Command        string
	AmountOfInk    int
	Shapetype      string
	ShapeSvgString string
	Fill           string
	Stroke         string
	PairKey        ecdsa.PrivateKey
	ValidFBlkNum	   uint8
	Opid		   uint32
}

func (o *Operation) CheckInk() bool {
	return false
}

func (o *Operation) CheckIntersection() bool {
	return false
}

func (o *Operation) CheckDuplicateSignature() bool {
	return false
}

func (o *Operation) CheckDeletedShapeExist() bool {
	return false
}

func (o *Operation) Validate() bool {
	// Check that each operation has sufficient ink associated with the public key that generated the operation.
	// Check that each operation does not violate the shape intersection policy described above.
	// Check that the operation with an identical signature has not been previously added to the blockchain (prevents operation replay attacks).
	// Check that an operation that deletes a shape refers to a shape that exists and which has not been previously deleted.
	return false
}
