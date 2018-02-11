// Artnode - Miner API

package am
import (
"net/rpc"
"crypto/ecdsa"
"fmt"
"time"
)

// type Operation struct {
// 	Opsig string }
// 	// add other fields of operation

type ArtNodeStruct struct {
	ArtNodeId  int
	PairKey    ecdsa.PrivateKey
	AmConn     *rpc.Client
}

type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32
	CanvasYMax uint32
}

type InitialCanvasSetting struct{
	Cs CanvasSettings
	ListOfOps_str []string
}

// var ArtNodeHeartBeat uint32; ArtNodeHeartBeat = 800
// Heartbeat from the Art node -- implements this interface
func (a *ArtNodeStruct) ArtNodeHeartBeat() {
	artNodeStatus := false
	for {
		error := a.AmConn.Call("HeartBeat.ArtNodeHeartBeat","hey", &artNodeStatus) // probally send the whole object
		if error != nil {
			fmt.Println(error)
		}
		time.Sleep(time.Millisecond * time.Duration(800))
		fmt.Println("ArtNodeHeartBeat(): Art node is alive")
	}
}

func (a *ArtNodeStruct) GetCanvasSettings () (InitialCanvasSetting, error) {
	initCS := &InitialCanvasSetting{}
	err := a.AmConn.Call("CanvasSet.GetCanvasSettingsFromMiner", "hey", initCS)
	CheckError(err)
	return *initCS, err
}

func (a *ArtNodeStruct) ArtnodeOp (s string) (error) {
	alive := false
	err := a.AmConn.Call("ArtNodeOpReg.DoArtNodeOp", "hey", &alive)
	CheckError(err)
	return nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
