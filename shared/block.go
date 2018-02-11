package shared

import (
	"crypto/ecdsa"
	"strconv"
)

type BlockApi interface {
	GetStringBlock() string
}

type Block struct {
	CurrentHash       string
	PreviousHash      string
	PreviousOPs       []Operation
	CurrentOP         Operation
	Children          []*Block
	DistanceToGenesis int
	Nonce             int32
	PublicKey         ecdsa.PublicKey
}

// Return a string repersentation of PreviousHash, op, op-signature, pub-key,
func (b Block) GetString() string {
	return b.CurrentHash + b.CurrentOP.Command + b.CurrentOP.UserSignature + pubKeyToString(b.PublicKey)

	// for _, operation := range b.PreviousOPs {
	// 	operationString := operation.Command + "," + operation.Shapetype + " by " + operation.UserSignature + " \n "
	//
	// 	nonce += operationString
	// }
	// ret += b.CurrentOP.Command
	// return nonce
}

func (b *Block) checkMD5() bool {
	if computeNonceSecretHash(b.GetString(), strconv.FormatInt(int64(b.Nonce), 10)) == b.CurrentHash {
		return true
	}
	return false
}

func (b *Block) checkValidOPsSig() bool {
	return false
}

func (b *Block) checkPreviousHash() bool {
	return false
}

func (b *Block) Validate() bool {
	// Check that the nonce for the block is valid: PoW is correct and has the right difficulty.
	// Check that each operation in the block has a valid signature (this signature should be generated using the private key and the operation).
	// Check that the previous block hash points to a legal, previously generated, block.
	return b.checkMD5() && b.checkValidOPsSig() && b.checkPreviousHash()
}
