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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"./blockartlib"
	// shared "./shared"
	//"encoding/gob"
	"encoding/gob"
)

var globalCanvas blockartlib.Canvas

func GetListOfOps(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hit the end point")
	s, err := json.Marshal(blockartlib.BlockChain)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	//Write json response back to response
	w.Write(s)
}

func ArtNodeAddshape(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}
		// TODO RPC call here
		globalCanvas.AddShape(2, blockartlib.PATH, "M 0 0 l 10 10", "transparent", "red")
		fmt.Println(string(body))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {
	gob.Register(&elliptic.CurveParams{})
	minerAddr := "127.0.0.1:39865" // hardcoded for now
	privKey := getKeyPair()        // TODO: use crypto/ecdsa to read pub/priv keys from a file argument.
	// Remove later
	minerAddrP := flag.String("ma", "MinerAddr Missing", "a string")
	minerPublicKey := flag.String("mp", "minerPublicKey missing", "a string")
	flag.Parse()
	fmt.Println("command line arguments ", *minerAddrP, " ", *minerPublicKey, minerAddr)

	// 	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(*minerAddrP, privKey)
	globalCanvas = canvas
	pieceOfShit := canvas.(blockartlib.CanvasObject)
	fmt.Printf("%+v", pieceOfShit.Ptr)
	fmt.Println("fuck this shit")

	artAddres := pieceOfShit.Ptr.ArtNodeipStr
	// port := artNodeIp.String()[strings.Index(artNodeIp.String(), ":"):len(artNodeIp.String())]
	port := artAddres[strings.Index(artAddres, ":"):len(artAddres)]
	fmt.Println(port)
	mux := http.NewServeMux()
	mux.HandleFunc("/getshapes", GetListOfOps)
	mux.HandleFunc("/addshape", ArtNodeAddshape)

	go http.ListenAndServe(":5000", mux)

	if checkError(err) != nil {
		fmt.Println(err.Error())
		return
	}
	//For testing,Can be deleted
	//  isOpvalid,testOp := canvas.IsSvgStringValid("m 100 100 l 500 400 l 1000 2000")
	//	isOutofBound := canvas.IsSvgOutofBounds(testOp)
	//	fmt.Println("operation first second third",isOpvalid,string(testOp.MovList[0].Cmd),string(testOp.MovList[1].Cmd))
	//	fmt.Println("Operation is out of bound!:",isOutofBound)

	validateNum := 2
	fmt.Println("remove after", canvas, settings, validateNum)

	// Getter method checks
	fmt.Println("remove after", canvas, settings, validateNum)
	ink, err := canvas.GetInk()
	fmt.Println("art-app.main(): going to get ink from miner", ink, "   ", err)
	gb, err := canvas.GetGenesisBlock()
	fmt.Println("art-app.main(): going to get genesis block from miner", gb, "   ", err)

	// Add a line.

	fmt.Println("ADDING SHAPES+++++")

	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 0 0 l 10 10", "transparent", "red")
	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 2 9 l 10 10", "transparent", "blue")
	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 20 90 l 10 10", "transparent", "green")
	// _, _, _, err = canvas.AddShape(2, blockartlib.PATH, "M 21 98 l 10 10", "transparent", "black")

	if checkError(err) != nil {
		return
	}

	for {

	}

	return

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
