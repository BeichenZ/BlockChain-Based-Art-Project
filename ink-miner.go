package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	shared "./shared"
)

func main() {
	servAddr := os.Args[1]
	publicKey := os.Args[2]
	privKey := os.Args[3]
	numberOfZeroes := 3 //This will be change if server.go is online
	guess := int64(0)   //This will be calculated
	var zeroesBuffer bytes.Buffer
	for i := int64(0); i < int64(numberOfZeroes); i++ {
		zeroesBuffer.WriteString("0")
	}
	zeroes := zeroesBuffer.String()

	inkMinerStruct := initializeMiner(servAddr, publicKey, privKey)
	fmt.Println(inkMinerStruct)
	// TODO register a miner node here, get back Neighbours info and threshold

	// TODO start heartbeat to the server

	if len(inkMinerStruct.Neighbours) > inkMinerStruct.Threshold {
		// TODO start Mining for no-op
		mine("changeThisLater", guess, numberOfZeroes, zeroes)
		// TODO break the mining loop once a secret is found
		// TODO flood the network
	}

	return
}

func initializeMiner(servAddr string, publicKey string, privKey string) shared.MinerStruct {
	return shared.MinerStruct{ServerAddr: servAddr, PublicKey: publicKey, PrivKey: privKey}
}

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	fmt.Println(str)
	return str
}

//this is a poor algorithm, feel free to change it
func mine(nonce string, guess int64, numberOfZeroes int, zeroes string) string {
	for {
		guessString := strconv.FormatInt(guess, 10)
		if computeNonceSecretHash(nonce, guessString)[32-numberOfZeroes:] == zeroes {
			fmt.Println(guessString)
			return guessString
		}
		guess++
	}
}
