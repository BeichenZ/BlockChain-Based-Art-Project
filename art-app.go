/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"./blockartlib"
	"fmt"
	"os"
	"flag"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	//"encoding/gob"
	)

func main() {
	minerAddr := "127.0.0.1:39865" // hardcoded for now
	privKey := getKeyPair()// TODO: use crypto/ecdsa to read pub/priv keys from a file argument.
	// Remove later
	minerAddrP := flag.String("ma", "MinerAddr Missing", "a string")
	minerPublicKey := flag.String("mp", "minerPublicKey missing", "a string")
	flag.Parse()
	fmt.Println("command line arguments ",*minerAddrP," ", *minerPublicKey, minerAddr)

// 	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(*minerAddrP, privKey)
	if checkError(err) != nil {
		fmt.Println("this is 37")
		return
	}

    validateNum := 2
    fmt.Println("remove after", canvas, settings, validateNum)
	// Add a line.
//	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	if checkError(err) != nil {
		return
	}

	// Add another line.
//	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	if checkError(err) != nil {
		return
	}

	// Delete the first line.
//	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if checkError(err) != nil {
		return
	}

	// assert ink3 > ink2

	// Close the canvas.
//	ink4, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}

// Helper functions

// gets the key pair given the public key of the miner --- change***
func getKeyPair() ecdsa.PrivateKey {
	artMinerKeyPair, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	return *artMinerKeyPair
}
