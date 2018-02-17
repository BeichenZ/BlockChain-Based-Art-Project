package shared

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"net/rpc"
	"time"
)

type Block_OP_Index struct {
	IndexB, IndexO int
}
type MinerRPCServer struct {
	Miner *MinerStruct
}

type BadBlockError string

func (e BadBlockError) Error() string {
	return fmt.Sprintf("BAD BLOCK")
}

func (m *MinerRPCServer) SendChain(s string, blockChain *BlockPayloadStruct) error {
	log.Println(s + "============================================================")
	thisMinerBlockChain := CopyBlockChainPayload(m.Miner.BlockChain)
	*blockChain = thisMinerBlockChain
	return nil
}

func (m *MinerRPCServer) StopMining(block *Block, alive *bool) error {
	log.Println("stopped")
	if !m.Miner.FoundHash {
		log.Print("Someone send me a block, its hash is: ", block.CurrentHash)
		if block.Validate() {

			go func() {
				m.Miner.MiningStopSig <- block
			}()
		} else {
			return BadBlockError("BAD BLOCK")
		}
	} else {
		log.Print("I have found the hash, but so did at least one other miner")
		AddNewBlock(m.Miner.BlockChain, block)
	}
	log.Println("Sent channel info")
	return nil
}

func (m *MinerRPCServer) ReceivedOperation(operation *Operation, alive *bool) error {

	m.Miner.RecievedOpSig <- *operation

	return nil
}

func (m *MinerRPCServer) ReceiveMinerHeartBeat(minerNeighbourAddr string, alive *bool) error {
	allNeighbour.Lock()
	defer allNeighbour.Unlock()
	// log.Println("Heartbeat received from ", minerNeighbourAddr)
	// fmt.Println("_____________________________________________________________________________________")
	// fmt.Printf("%+v", allNeighbour.all)
	if _, ok := allNeighbour.all[minerNeighbourAddr]; ok {
		// log.Println("The miner is in the map")
		allNeighbour.all[minerNeighbourAddr].RecentHeartbeat = time.Now().UnixNano()
		// fmt.Println("Heartbeat RECEIVED ", allNeighbour.all[minerNeighbourAddr].RecentHeartbeat)
	} else {
		// log.Println("Nothing in the map")
	}
	return nil
}

func (m *MinerRPCServer) ActivateHeartBeat(SenderAddr string, alive *bool) error {
	go m.Miner.minerSendHeartBeat(SenderAddr)
	return nil
}

func (m *MinerRPCServer) ArtNodeRegister(ArtNodeAddr string, reply *bool) error {
	if _, exist := allArtNodes.all[ArtNodeAddr]; exist {
		fmt.Println("Artnode already here")
	} else {
		allArtNodes.Lock()
		allArtNodes.all[ArtNodeAddr] = 1
		allArtNodes.Unlock()
	}
	return nil
}

func (m *MinerRPCServer) MinerRegister(MinerNeighbourPayload *string, thisMinerChainLength *int) error {
	fmt.Println("------------------------------------I got here--------------------------------------")
	fmt.Println(MinerNeighbourPayload)
	// if len(allNeighbour.all) > int(m.Miner.Settings.MinNumMinerConnections) {
	// 	m.Miner.NotEnoughNeighbourSig <- false
	// }

	if _, ok := allNeighbour.all[(*MinerNeighbourPayload)]; ok {
		log.Println("The Miner is already here")
	} else {
		alive := false
		number := 0
		allNeighbour.all[(*MinerNeighbourPayload)] = &MinerStruct{
			MinerAddr: (*MinerNeighbourPayload),
			// MinerConnection: &MinerNeighbourPayload.client,
			RecentHeartbeat: time.Now().UnixNano(),
		}
		_, length := findDeepestBlocks(m.Miner.BlockChain, 0)
		//length := m.Miner.FindLongestChainLength()
		*thisMinerChainLength = length
		log.Println("Registration time is: ", time.Now().UnixNano())
		log.Println("Successfully recorded this neighbouring miner", (*MinerNeighbourPayload))
		client, error := rpc.Dial("tcp", *MinerNeighbourPayload)
		if error != nil {
			fmt.Println(error)
		}
		err := client.Call("MinerRPCServer.MinerRegister", m.Miner.MinerAddr, &number)
		if err != nil {
			fmt.Println(err)
		}
		go monitor(*MinerNeighbourPayload, *m.Miner, 1000*time.Millisecond)

		client.Call("MinerRPCServer.ActivateHeartBeat", m.Miner.MinerAddr, &alive)

	}
	return nil
}

type KeyCheck int

func (l *KeyCheck) ArtNodeKeyCheck(privKey *string, reply *bool) error {
	*reply = true
	fmt.Println("ArtNodeKeyCheck(): Art node connecting with me")

	return nil
}

type CanvasSet struct {
	Miner MinerStruct
}
type ArtNodeOpReg struct {
	Miner *MinerStruct
}

func (l *ArtNodeOpReg) DoArtNodeOp(op *Operation, reply *int) error {
	//reply decode: 0->Success,1->insufficientInk, 2->OverlappedShape
	// Check InsufficientInkError
	fmt.Printf("%+v", l.Miner)
	if l.Miner.MinerInk < (*op).AmountOfInk {
		fmt.Println("The current ink we have is:", l.Miner.MinerInk)
		fmt.Println("Insufficient Ink Detected for shape:", (*op).Command, "With requested ink:", (*op).AmountOfInk)
		*reply = 1
		return nil
	}
	// Check ShapeOverlapError
	/*
		if isOverLap := IsShapeOverLapWithOthers(op,l); isOverLap {
			fmt.Println("OverLapped Shape for shape string:", (*op).Command)
			*reply = 2
			return nil
		}
	*/
	fmt.Println(op.Command)

	// Sign the Operation
	r, s, err := ecdsa.Sign(rand.Reader, &l.Miner.PairKey, []byte((*op).ShapeSvgString))

	if err != nil {
		fmt.Println(err)
	}

	op.Issuer = &l.Miner.PairKey
	op.IssuerR = r
	op.IssuerS = s
	fmt.Println(pubKeyToString(op.Issuer.PublicKey))
	go func() {
		l.Miner.RecievedArtNodeSig <- *op
	}()

	blockCounter.Lock()
	origCounter := blockCounter.counter
	blockCounter.Unlock()

	// TODO  What if multiple art node tries to make operation

	validNum := op.ValidFBlkNum
	// ** Maybe set timeout if it takes too long
	// while l.Miner.chain is not validateNum more than before don't return anything, only return when its not
	waitStartTime := time.Now()
	then := time.Now()
	for {
		//fmt.Println("DoArtNodeOp: IN loop")
		//Set-up overall timeout
		then = time.Now()
		if then.Sub(waitStartTime).Seconds() > 10 {
			*reply = 3
			return DisconnectedError("Wait For ValidateNum Take Too Long")
		}
		blockCounter.Lock()
		if (blockCounter.counter - origCounter) == validNum {
			*reply = 0
			blockCounter.Unlock()
			break
		}
		blockCounter.Unlock()
	}
	fmt.Println("DoArtNodeOp() validateNum condition satisfied")
	return nil
}

func IsShapeOverLapWithOthers(op *Operation, l *ArtNodeOpReg) bool {
	//For operation from same miner , do not check
	//For operation from different miner, check for overlapping
	/*
		svgString := (*op).ShapeSvgString
		newsvgFill := (*op).Fill
		svgStroke := (*op).Stroke
		_, newSvgOp := IsSvgStringParsable_Parse(svgString)
		newSvgOp_IsClosed, newvtxArr, newedgeArr := IsClosedShapeAndGetVtx(newSvgOp)
		var tarSvgOp SingleOp
		var tarSvgOp_IsClosed bool
		var tarvtxArr []Point
		var taredgeArr []LineSectVector
		isTwoOverLap := false
		//Below Two varialbe is to prevent the siutation of add->delete->false conflict situation
		var Block_OP_IndexArr []Block_OP_Index //Record all the conflicting shape and check if they have been deleted
		var BlockOpIndexPair_AppearCount []int //Store number of such pair happens, No one should appear even numbers

		minerCopy := *(l.Miner)
		longestChainArr_Invt := getLongestPath(minerCopy.BlockChain)
		//Loop through Each Block.
		for indexB, tarBlock := range longestChainArr_Invt {
			//Loop through Block.Operation
			for indexO, tarOp := range tarBlock.CurrentOPs {
				isTwoOverLap := false //Initialize to be false for each check
				_, tarSvgOp = IsSvgStringParsable_Parse(tarOp.ShapeSvgString)
				tartSvgOp_IsClosed, tarvtxArr, taredgeArr := IsClosedShapeAndGetVtx(tarSvgOp)
				//Discuss By Case,If No OverLap,
				//Case: Both the shape are transparent (line only) shape
				if newsvgFill == "transparent" && tarOp.Fill == "transparent" {
					isOverLap= IsTwoLineShapeOverLap(newvtxArr,newedgeArr,tarvtxArr,taredgeArr)
				} else if  (newsvgFill == "transparent" && tar
				//Case:One Line type, One Fill Type
				//Case:Two Fill Type
			}

		}

		//Check if the any of the conflicting shape has actually be deleted
	*/
	return true
}

func (l *ArtNodeOpReg) ArtnodeInkRequest(s string, remainInk *uint32) error {
	*remainInk = l.Miner.GetInkBalance()
	return nil
}

func (l *ArtNodeOpReg) ArtnodeGenBlkRequest(s string, genBlkHash *string) error {
	*genBlkHash = l.Miner.BlockChain.CurrentHash
	fmt.Println("ArtnodeGenBlkRequest() ", genBlkHash)
	return nil
}

func (l *ArtNodeOpReg) ArtnodeBlkChildRequest(bHash string, blkCh *[]string) (err error) {
	*blkCh, err = l.Miner.GetBlkChildren(l.Miner.BlockChain, bHash)
	return err
}

func (l *ArtNodeOpReg) ArtnodeSvgStringRequest(shapeHash string, svgString *string) (err error) {
	*svgString = l.Miner.GetSVGShapeString(l.Miner.BlockChain, shapeHash)
	return err
}
func (l *ArtNodeOpReg) ArtnodeGetOpWithHashRequest(shapeHash string, opToDel *Operation) error {
	*opToDel = l.Miner.GetOpToDelete(l.Miner.BlockChain, shapeHash)
	return nil
}

func (l *CanvasSet) GetCanvasSettingsFromMiner(s string, ics *InitialCanvasSetting) error {
	fmt.Println("request for CanvasSettings")
	ics.Cs = CanvasSettings(l.Miner.Settings.CanvasSettings)
	ics.ListOfOps_str = l.Miner.ListOfOps_str
	l.Miner.setUpConnWithArtNode(s)
	fmt.Println("GetCanvasSettingsFromMiner() ", *ics)
	return nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
