/*

This package specifies the application's interface to the the BlockArt
library (blockartlib) to be used in project 1 of UBC CS 416 2017W2.

*/

package blockartlib

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"strings"
	"net/rpc"
	am "./artminerlib"
	
)

// Represents a type of shape in the BlockArt system.
type ShapeType int

const (
	// Path shape.
	PATH ShapeType = iota

	// Circle shape (extra credit).
	// CIRCLE
)

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32
	CanvasYMax uint32
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string

	// The minimum number of ink miners that an ink miner should be
	// connected to. If the ink miner dips below this number, then
	// they have to retrieve more nodes from the server using
	// GetNodes().
	MinNumMinerConnections uint8

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32
	InkPerNoOpBlock uint32

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8
	PoWDifficultyNoOpBlock uint8

	// Canvas settings
	canvasSettings CanvasSettings
}

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains address IP:port that art node cannot connect to.
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("BlockArt: cannot connect to [%s]", string(e))
}

// Contains amount of ink remaining.
type InsufficientInkError uint32

func (e InsufficientInkError) Error() string {
	return fmt.Sprintf("BlockArt: Not enough ink to addShape [%d]", uint32(e))
}

// Contains the offending svg string.
type InvalidShapeSvgStringError string

func (e InvalidShapeSvgStringError) Error() string {
	return fmt.Sprintf("BlockArt: Bad shape svg string [%s]", string(e))
}

// Contains the offending svg string.
type ShapeSvgStringTooLongError string

func (e ShapeSvgStringTooLongError) Error() string {
	return fmt.Sprintf("BlockArt: Shape svg string too long [%s]", string(e))
}

// Contains the bad shape hash string.
type InvalidShapeHashError string

func (e InvalidShapeHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid shape hash [%s]", string(e))
}

// Contains the bad shape hash string.
type ShapeOwnerError string

func (e ShapeOwnerError) Error() string {
	return fmt.Sprintf("BlockArt: Shape owned by someone else [%s]", string(e))
}

// Empty
type OutOfBoundsError struct{}

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("BlockArt: Shape is outside the bounds of the canvas")
}

// Contains the hash of the shape that this shape overlaps with.
type ShapeOverlapError string

func (e ShapeOverlapError) Error() string {
	return fmt.Sprintf("BlockArt: Shape overlaps with a previously added shape [%s]", string(e))
}

// Contains the invalid block hash.
type InvalidBlockHashError string

func (e InvalidBlockHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid block hash [%s]", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a canvas in the system.
type Canvas interface {
	// Adds a new shape to the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - InsufficientInkError
	// - InvalidShapeSvgStringError
	// - ShapeSvgStringTooLongError
	// - ShapeOverlapError
	// - OutOfBoundsError
	AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error)

	// Returns the encoding of the shape as an svg string.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidShapeHashError
	GetSvgString(shapeHash string) (svgString string, err error)

	// Returns the amount of ink currently available.
	// Can return the following errors:
	// - DisconnectedError
	GetInk() (inkRemaining uint32, err error)

	// Removes a shape from the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - ShapeOwnerError
	DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error)

	// Retrieves hashes contained by a specific block.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetShapes(blockHash string) (shapeHashes []string, err error)

	// Returns the block hash of the genesis block.
	// Can return the following errors:
	// - DisconnectedError
	GetGenesisBlock() (blockHash string, err error)

	// Retrieves the children blocks of the block identified by blockHash.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetChildren(blockHash string) (blockHashes []string, err error)

	// Closes the canvas/connection to the BlockArt network.
	// - DisconnectedError
	CloseCanvas() (inkRemaining uint32, err error)
}

// The constructor for a new Canvas object instance. Takes the miner's
// IP:port address string and a public-private key pair (ecdsa private
// key type contains the public key). Returns a Canvas instance that
// can be used for all future interactions with blockartlib.
//
// The returned Canvas instance is a singleton: an application is
// expected to interact with just one Canvas instance at a time.
//
// Can return the following errors:
// - DisconnectedError
func OpenCanvas(minerAddr string, privKey ecdsa.PrivateKey) (canvas Canvas, setting CanvasSettings, err error) {
	// TODO
	fmt.Print("OpenCanvas(): Going to connect to miner")

	// Connect to Miner
	art2MinerCon, err := rpc.Dial("tcp", minerAddr)
	CheckError(err)
	fmt.Println("Connected  to Miner")

	// see if the Miner key matches the one you have
	var reply bool
	Key := "test"
	err = art2MinerCon.Call("KeyCheck.ArtNodeKeyCheck", Key, &reply)
	CheckError(err)
	if reply {
		fmt.Println("ArtNode has same key as miner")
	var thisCanvasObj CanvasObject
	 thisCanvasObj.ArtNode.AmConn = art2MinerCon

	 // Art node gets canvas settings from Miner node
	 fmt.Println("ArtNode going to get settings from miner")
	 // old 
	 initMs, err := thisCanvasObj.ArtNode.GetCanvasSettings() // get the canvas settings, list of current operations
	 thisCanvasObj.ListOfOps = initMs.ListOfOps
	 setting = CanvasSettings(initMs.Cs)
	 CheckError(err)

	return thisCanvasObj, setting, nil
	} else { fmt.Println("ArtNode does not have same key as miner")
		return nil, CanvasSettings{}, DisconnectedError("")  }
 	
}

//Implementation of Canvas Interface
type CanvasObject struct{
	ArtNode am.ArtNodeStruct
	ListOfOps []string
	// Canvas settings field?
}  

func (t CanvasObject) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	// check if there's enough ink for the operation
	// send operation to the miner Call()
	//t.ArtNode.AmConn.doOp(s string)
	//Check for ShapeSvgStringTooLongError
	if len(shapeSvgString) > 128 {
		return "", "", 0, ShapeSvgStringTooLongError(shapeSvgString)
	}
	if !t.IsSvgStringValid(shapeSvgString) {
		return "", "", 0, InvalidShapeSvgStringError(shapeSvgString)
	}
	if !t.IsSvgOutofBounds(shapeSvgString) {
		return "", "", 0, OutOfBoundsError{}
	}
	return "", "", 0, nil
}
func (t CanvasObject) GetSvgString(shapeHash string) (svgString string, err error) {
	// TODO
	return "", nil
	
}
func (t CanvasObject) GetInk() (inkRemaining uint32, err error) {
	// TODO
	// get longest branch from miner compute ink based on how many signitures are from the miner
	return 0, nil
}
func (t CanvasObject) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	// TODO
	return 0, nil
}
func (t CanvasObject) GetShapes(blockHash string) (shapeHashes []string, err error) {
	var s []string
	return s, nil
	}
func (t CanvasObject) GetGenesisBlock() (blockHash string, err error) {
	// TODO
	// Request block chain from miner 
	return "", nil
}
func (t CanvasObject) GetChildren(blockHash string) (blockHashes []string, err error) {
	var s []string
	return s, nil
}
func (t CanvasObject) CloseCanvas() (inkRemaining uint32, err error){
	// TODO
	return 0, nil
	
}
// Added helpers
func (t CanvasObject) IsSvgStringValid(shapeSvgString string) bool {
	//To Be Implemented
	availableCmds := []byte{77, 109, 86, 118, 72, 104, 76, 108, 90, 122} // MmVvLlZz
	svgCharArray := []byte(strings.TrimSpace(shapeSvgString))

	if !bytes.Contains(availableCmds, svgCharArray[0:0]) {
		return false
	} else {

	}

	return false
}
func (t CanvasObject) IsSvgOutofBounds(shapeSvgString string) bool {
	//To be Implemented
	return false
}

// Additional Helper Functions
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

