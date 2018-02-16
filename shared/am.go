// Artnode - Miner API

package shared
import (
"net/rpc"
"crypto/ecdsa"
"fmt"
//"time"
)

type ArtNodeStruct struct {
	ArtNodeId  int
	PairKey    ecdsa.PrivateKey
	AmConn     *rpc.Client
}

type InitialCanvasSetting struct{
	Cs CanvasSettings
	ListOfOps_str []string
}
type AddShapeReply struct {
	Success bool
	Err error
}
func (a *ArtNodeStruct) GetCanvasSettings () (InitialCanvasSetting, error) {
	initCS := &InitialCanvasSetting{}
	err := a.AmConn.Call("CanvasSet.GetCanvasSettingsFromMiner", "hey", initCS)
	CheckError(err)
	return *initCS, err
}
// Artnode issues an operation
// returns a number which indicates which indicates the status of the operation
		// for now boolean
func (a *ArtNodeStruct) ArtnodeOp (op Operation) (validOp bool,err error) {
	var reply int
	locaerr := a.AmConn.Call("ArtNodeOpReg.DoArtNodeOp", op, &reply)
	CheckError(locaerr)
	//TODO:Parse the return int to error
	return false,InvalidShapeSvgStringError("haha")
}

func (a *ArtNodeStruct) GetInkBalFromMiner() (uint32, error) {
	var i uint32
	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeInkRequest", "ink pls", &i)
	fmt.Println("GetInkBalFromMiner() ", i)
	return i, err
}
func (a *ArtNodeStruct) GetGenesisBlockFromMiner() (string, error) {
	var gb string
	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeGenBlkRequest", "Genisis blk", &gb)
	return gb, err
}
func (a *ArtNodeStruct) GetChildrenFromMiner(bHash string) ([]string, error)  {
	var mch []string

	err:= a.AmConn.Call("ArtNodeOpReg.ArtnodeBlkChildRequest", bHash, &mch)
	return mch, err
}
func (a *ArtNodeStruct) GetSvgStringUsingOperationSignature(shapeHash string) (string, error) {
	var svgstring string

	err := a.AmConn.Call("ArtNodeOpReg.ArtnodeSvgStringRequest", shapeHash, &svgstring)
	return svgstring, err
}
