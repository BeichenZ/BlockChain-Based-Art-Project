/*

This package specifies the application's interface to the the BlockArt
library (blockartlib) to be used in project 1 of UBC CS 416 2017W2.

*/

package blockartlib

import (
	//"bytes"
	"crypto/ecdsa"
	"fmt"
	//"strings"
	"net/rpc"
	"regexp"
	shared "../shared"
	"strconv"
	//"time"
	"math/rand"
	"time"
	"math"
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

	//Testing functions, to be deleted
	/*
	IsSvgStringValid(svgStr string) (isValid bool,Op shared.SingleOp)
  IsSvgOutofBounds(svgOP shared.SingleOp) bool
  */
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
	fmt.Println("GOING TO CALL ARTNODEKEYCHECK")
	err = art2MinerCon.Call("KeyCheck.ArtNodeKeyCheck", Key, &reply)
	CheckError(err)
	if reply {
		fmt.Println("ArtNode has same key as miner")
	var thisCanvasObj CanvasObject
	 thisCanvasObj.ptr = new(CanvasObjectReal)
	 thisCanvasObj.ptr.ArtNode.AmConn = art2MinerCon

	 // Art node gets canvas settings from Miner node
	 fmt.Println("ArtNode going to get settings from miner")
	 // old
	 initMs, err := thisCanvasObj.ptr.ArtNode.GetCanvasSettings() // get the canvas settings, list of current operations
	 setting = CanvasSettings(initMs.Cs)
	 thisCanvasObj.ptr.ListOfOps_str = initMs.ListOfOps_str
	 thisCanvasObj.ptr.XYLimit = shared.Point{X:float64(setting.CanvasXMax),Y:float64(setting.CanvasYMax)}

	 CheckError(err)

	return thisCanvasObj, setting, nil
	} else { fmt.Println("ArtNode does not have same key as miner")
		return nil, CanvasSettings{}, DisconnectedError("")  }

}

//Underlying struct holding the info for canvas
type CanvasObjectReal struct{
	ArtNode shared.ArtNodeStruct
	ListOfOps_str []string
	ListOfOps_ops []shared.SingleOp
	LastPenPosition shared.Point
	XYLimit				shared.Point

	// Canvas settings field?
}

type CanvasObject struct {
	ptr *CanvasObjectReal
}

func (t CanvasObject) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	// check if there's enough ink for the operation
	// send operation to the miner Call()
	//time.Sleep(20000 * time.Millisecond)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	randNum := r1.Intn(100)
	fmt.Println("AddShape(): The command is draw thing " + string(randNum) )
	newOP := shared.Operation{
		Command:"draw things" + string(randNum),
		ValidFBlkNum: validateNum,
		Opid: rand.Uint32(),
	}
	validOp:=t.ptr.ArtNode.ArtnodeOp(newOP) // fn needs to return boolean


	//
	//for {
	//	if condition {
	//		break
	//	}
	//
	//}
	fmt.Println("AddShape() ", validOp)
	//Check for ShapeSvgStringTooLongError
	var svgOP shared.SingleOp
	if len(shapeSvgString) > 128 {
		return "", "", 0, ShapeSvgStringTooLongError(shapeSvgString)
	}
	valid,svgOP :=t.IsSvgStringValid(shapeSvgString);
	if !valid {
		return "", "", 0, InvalidShapeSvgStringError(shapeSvgString)
	}
	if !t.IsSvgOutofBounds(svgOP) {
		return "", "", 0, OutOfBoundsError{}
	}

	//Once successfully add shape. Finish Post Settings
		//1.update LastPenPosition
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
func (t CanvasObject) IsSvgStringValid(svgStr string) (isValid bool,Op shared.SingleOp){
	//To Be Implemented
	//Legal Example: "m 20 0 L 19 21",all separated by space,always start at m/M
	strCnt := len(svgStr)
	var movList []shared.SingleMov
	parsedOp := shared.SingleOp{MovList:movList}
	if strCnt < 3 {
		return false,parsedOp
	}
	if svgStr[0] != 'M'{
		return false,parsedOp
	}
	regex_2 := regexp.MustCompile("([mMvVlLhHZz])[\\s]([-]*[0-9]+)[\\s]([-]*[0-9]+)")
	regex_1 := regexp.MustCompile("([mMvVlLhHZz])[\\s]([-]*[0-9]+)")
  var matches []string
	var oneMov shared.SingleMov
	var thisRune rune
	fmt.Println("string size is",strCnt)
	for i := 0; i < strCnt; i = i {
		fmt.Println("Current I is",i)
		thisRune = rune(svgStr[i])
		switch thisRune {
		case 'm', 'M', 'L', 'l':
			arr := regex_2.FindStringIndex(svgStr[i:])
			if arr == nil {
				return false,parsedOp
			} else {
				//if legal, Parse it
				matches = regex_2.FindStringSubmatch(svgStr[i:])
				intVal1,_ := strconv.Atoi(matches[2])
				intVal2,_ := strconv.Atoi(matches[3])
				oneMov = shared.SingleMov{Cmd:rune(thisRune),X:float64(intVal1),Y:float64(intVal2),ValCnt:2}
				parsedOp.MovList = append(parsedOp.MovList,oneMov)
				//Update Index
				fmt.Println("ML update next index is",arr[0],arr[1])
				i = i+arr[1]+1
			}
		case 'v', 'V', 'H', 'h':
			arr := regex_1.FindStringIndex(svgStr[i:])
			fmt.Println("VH update next index is",arr[0],arr[1])
			if arr == nil {
				return false,parsedOp
			} else {
				matches = regex_1.FindStringSubmatch(svgStr[i:])
				intVal1,_ := strconv.Atoi(matches[2])
				oneMov = shared.SingleMov{Cmd:rune(thisRune),X:float64(intVal1),Y:0,ValCnt:1}
				parsedOp.MovList = append(parsedOp.MovList,oneMov)
				i = i+arr[1]+1
			}
		case 'Z','z':
			oneMov := shared.SingleMov{Cmd:rune(thisRune),ValCnt:0}
			parsedOp.MovList = append(parsedOp.MovList,oneMov)
			i = i + 2
		default:
			return false,parsedOp
		}
	}
	return true,parsedOp//pass all tests
}
func (t CanvasObject) IsSvgOutofBounds(svgOP shared.SingleOp) bool {
	xVal := t.ptr.LastPenPosition.X
	yVal := t.ptr.LastPenPosition.Y
	for _,element := range svgOP.MovList {
		switch element.Cmd {
		case 'M','L','H','V':
			if element.X > t.ptr.XYLimit.X || element.X < 0 || element.Y > t.ptr.XYLimit.Y || element.Y < 0{
				return true
			} else {
					xVal = xVal + element.X
					yVal = yVal + element.Y
			}
		case 'm','l','v','h':
			if element.X+xVal > t.ptr.XYLimit.X || element.X+xVal < 0 || element.Y+yVal > t.ptr.XYLimit.Y || element.Y+yVal < 0 {
				return true
			} else {
				xVal = xVal + element.X
				yVal = yVal + element.Y
			}
		default:
		}
	}
	return false
}
func (t CanvasObject) ParseOpsStrings(){
	opsArrSize := len(t.ptr.ListOfOps_ops)
	for i,element := range t.ptr.ListOfOps_str {
		if valid,oneOp := t.IsSvgStringValid(element);valid{
			if i <= (opsArrSize-1){
				t.ptr.ListOfOps_ops[i] = oneOp
			} else {
				t.ptr.ListOfOps_ops=append(t.ptr.ListOfOps_ops,oneOp)
			}
		}
	}
}
func (t CanvasObject) IsClosedShapeAndGetVtx(op shared.SingleOp) (IsClosed bool,vtxArray []shared.Point,edgeArray []shared.LineSectVector) {
	var vtxArr []shared.Point
	var edgeArr   []shared.LineSectVector
	var curVtx    shared.Point
	var preVtx 		shared.Point
	var nextVtx   shared.Point
	var originalStart shared.Point
	var lastSubPathStart shared.Point
	//traverse all operation, identify list of edge and points
	//TODO : Corner Case when an open shape has the same ending point
	movCount := len(op.MovList)
	if movCount < 1 {return false,vtxArr,edgeArr}//Panic Check
	originalStart.X = op.MovList[0].X // Assume the first mov is always 'M' which is validated by IsValidSvgString
	originalStart.Y = op.MovList[0].Y
	for _,element := range op.MovList {
		switch element.Cmd {
		case 'M','V','H','L':
			preVtx = curVtx
			curVtx.X = element.X
			curVtx.Y = element.Y
			if element.Cmd != 'M'{
				edgeArr = append(edgeArr,shared.LineSectVector{preVtx,curVtx})//add new line segment
			}else {
				lastSubPathStart = curVtx // prepare for potential Z/z command
			}
			vtxArr = append(vtxArr,curVtx) // add new vertex
		case 'm','v','h','l':
			preVtx = curVtx
			curVtx.X += element.X
			curVtx.Y += element.Y
			if element.Cmd != 'm' {
				edgeArr = append(edgeArr,shared.LineSectVector{preVtx,curVtx})
			} else {
				lastSubPathStart = curVtx
			}
			vtxArr = append(vtxArr,curVtx)
		case 'Z','z':
			preVtx = curVtx
			curVtx = lastSubPathStart
			edgeArr = append(edgeArr,shared.LineSectVector{preVtx,curVtx})
		}
	}
	//List through the edge array and identify if everything is connected.Reuse variables
  if len(edgeArr) < 1 {return false, vtxArr,edgeArr}
	preVtx = edgeArr[0].Start
	for _,element := range edgeArr { // Check For discontinuity
		curVtx = element.Start
		nextVtx = element.End
	  if curVtx != preVtx{ return false,vtxArr,edgeArr} else {
			vtxArr = append(vtxArr,curVtx)
			vtxArr = append(vtxArr,nextVtx)
			preVtx = nextVtx
		}
	}
	//If entire edge is continous, Check if it returns to the same points
	if nextVtx != originalStart {
		return false,vtxArr,edgeArr
		} else {
		uniqueVtxCount := len(vtxArr)-1
		return true,vtxArr[:uniqueVtxCount],edgeArr // the last "nextVtx" will be an overlapping of the staring point
	}
}
func (t CanvasObject) CalculateShapeArea(IsClosed bool, vtxArr []shared.Point,edgeArr []shared.LineSectVector) float64 {
	//Given the parsed results from IsClosedShapeAndGetVtx
	var area float64
	area 	= float64(0)
	if !IsClosed {
		//Non-Closed Shape's Area is the summation of the
		for _,element := range edgeArr {
			 area += Distance_TwoPoint(element.Start,element.End)
		}
	} else {
		//TODO: Check if it is  self intersecting and calculate area accordingly

	}
	return area
}
//Input should be already checked as Closed shape
func (t CanvasObject) IsSelfIntersect_GetSingleShapes(vtxArr []shared.Point) (IsSelfIntersect bool){
	//TODO: Check for self intersection and parse for individual shapes
	return false
}
//Geometric Function Functions
func Distance_TwoPoint(x,y shared.Point) float64 {
	return math.Sqrt(math.Pow(x.X-y.X,2)+math.Pow(x.Y-y.Y,2))
}
//Only Handle Non-Intersecting Polygon. Self-Intersected shape should be sub divided into before calling
func Area_SingleClosedPolygon(vtxArr []shared.Point) float64 {
	//Algorithm Reference:https://www.mathopenref.com/coordpolygonarea.html
  vtxCount := len(vtxArr)
	if vtxCount <= 1 { return 0}
	firstVtx := vtxArr[0]
	var nextVtx shared.Point
	area := float64(0)
	for index,element := range vtxArr {
		if index >= vtxCount-1 { nextVtx = firstVtx } else {
			nextVtx = vtxArr[index+1]
		}
		area += 0.5*(element.X*nextVtx.Y-element.Y*nextVtx.X)
	}
	return math.Abs(area)
}

// Additional Helper Functions
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
