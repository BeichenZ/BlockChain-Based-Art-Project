package shared

import (
	"crypto/ecdsa"
	"math/big"
	"fmt"
)

type Operation struct {
	Command        string
	AmountOfInk    uint32
	Shapetype      string
	ShapeSvgString string
	Fill           string
	Stroke         string
	Issuer        *ecdsa.PrivateKey
	IssuerR       *big.Int
	IssuerS       *big.Int
	ValidFBlkNum   uint8
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

func (o *Operation) CheckIssuerSig() bool {
	fmt.Println("CHECKING")
	fmt.Println( "ISSUER ", o.Issuer )
	fmt.Println( "ISSUERR ", o.IssuerR )
	fmt.Println( "ISSUES ", o.IssuerS )
	if (o.Issuer == nil) || ((o.IssuerR == nil) || (o.IssuerS == nil)){
		fmt.Println("------------------------------------------------They are all empty")

		return false
	}

	if ecdsa.Verify(&o.Issuer.PublicKey, []byte(o.Command), o.IssuerR, o.IssuerS){
		fmt.Println("----------------------------CORRECT OPERATION ISSUER SIGN")
	} else {
		fmt.Println("----------------------------INCORRECT OPERATION ISSUER SIGN")

	}



	return ecdsa.Verify(&o.Issuer.PublicKey, []byte(o.Command), o.IssuerR, o.IssuerS)
}
