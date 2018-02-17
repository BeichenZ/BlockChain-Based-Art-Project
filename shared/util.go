package shared

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"
)

type BlockDepthpair struct {
	block, depth interface{}
}

func monitor(minerNeighbourAddr string, miner MinerStruct, heartBeatInterval time.Duration) {
	for {
		allNeighbour.Lock()

		if time.Now().UnixNano()-allNeighbour.all[minerNeighbourAddr].RecentHeartbeat > int64(heartBeatInterval) {
			log.Printf("%s timed out, walalalalala\n", allNeighbour.all[minerNeighbourAddr].MinerAddr)
			delete(allNeighbour.all, minerNeighbourAddr)

			if len(allNeighbour.all) < int(miner.Settings.MinNumMinerConnections) {
				miner.NotEnoughNeighbourSig <- true
			}
			allNeighbour.Unlock()

			return
		}
		log.Printf("%s is alive\n", allNeighbour.all[minerNeighbourAddr].MinerAddr)
		allNeighbour.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func filter(m *MinerStruct, visited *[]*MinerStruct) bool {
	for _, s := range *visited {
		if s.MinerAddr == m.MinerAddr {
			return false
		}
	}
	return true
}

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

func doProofOfWork(m *MinerStruct, nonce string, numberOfZeroes int, newOPs []Operation, leadingBlock *Block, isDoingWorkForNoOp bool) *Block {
	i := int64(0)

	var zeroesBuffer bytes.Buffer
	for i := int64(0); i < int64(numberOfZeroes); i++ {
		zeroesBuffer.WriteString("0")
	}

	zeroes := zeroesBuffer.String()

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++Begin Proof of work+++++++++++++++++++++++++++++")
	for {
		select {
		case recievedBlock := <-m.MiningStopSig:
			fmt.Println("Received block from another miner")
			return recievedBlock
		case opFromMineNode := <-m.RecievedOpSig:
			fmt.Println("M-UPDATED OPERATION LIST FROM MINERS ")
			if isDoingWorkForNoOp {
				nonce = leadingBlock.CurrentHash + opFromMineNode.ShapeSvgString +opFromMineNode.Fill + opFromMineNode.Stroke + fmt.Sprint(opFromMineNode.AmountOfInk) + pubKeyToString(m.PairKey.PublicKey)
				newOPs = []Operation{opFromMineNode}
				isDoingWorkForNoOp = false
				fmt.Println(nonce)
				fmt.Println("M-Was calculating No-op, now calculating Operation ", opFromMineNode.Command)
			} else {
				m.OPBuffer = append(m.OPBuffer, opFromMineNode)
				fmt.Println("M-UPDATED OPERATION LIST FROM MINERS " + AllOperationsCommands(m.OPBuffer))
			}
		case opFromArtnode := <-m.RecievedArtNodeSig:
			if isDoingWorkForNoOp {
				isDoingWorkForNoOp = false
				nonce = leadingBlock.CurrentHash + opFromArtnode.ShapeSvgString + opFromArtnode.Fill + opFromArtnode.Stroke + fmt.Sprint(opFromArtnode.AmountOfInk) + pubKeyToString(m.PairKey.PublicKey)
				newOPs = []Operation{opFromArtnode}
				fmt.Println(nonce)
				fmt.Println("A-Was calculating No-op, now calculating Operation ", opFromArtnode.Command)
			} else {
				m.OPBuffer = append(m.OPBuffer, opFromArtnode)
			}
			visitedMiners := []*MinerStruct{m}
			fmt.Println("A-UPDATED OPERATION LIST FROM ART NODE" + AllOperationsCommands(m.OPBuffer))
			m.FloodOperation(&opFromArtnode, &visitedMiners)
		default:
			guessString := strconv.FormatInt(i, 10)

			hash := computeNonceSecretHash(nonce, guessString)
			if hash[32-numberOfZeroes:] == zeroes {
				log.Println("Found the hash, it is: ", hash)
				log.Println(" NOUNCE IS " + nonce)
				m.FoundHash = true
				return m.produceBlock(hash, newOPs, leadingBlock, guessString)
			}
			i++
		}
	}
}

func pubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}

func ParseBlockChain(thisBlock BlockPayloadStruct) *Block {
	x, y := elliptic.Unmarshal(elliptic.P384(), []byte(thisBlock.SolverPublicKey))
	if thisBlock.PreviousHash == "" {
		x = &big.Int{}
		y = &big.Int{}
	}
	fmt.Println(thisBlock.SolverPublicKey)
	fmt.Println(x)
	fmt.Println(y)
	producedBlock := &Block{
		CurrentHash:       thisBlock.CurrentHash,
		PreviousHash:      thisBlock.PreviousHash,
		R:                 &thisBlock.R,
		S:                 &thisBlock.S,
		CurrentOPs:        thisBlock.CurrentOPs,
		DistanceToGenesis: thisBlock.DistanceToGenesis,
		Nonce:             thisBlock.Nonce,
		SolverPublicKey: &ecdsa.PublicKey{
			Curve: elliptic.P384(),
			X:     x,
			Y:     y,
		},
	}
	var producedBlockChilden []*Block
	for _, child := range thisBlock.Children {
		producedBlockChilden = append(producedBlockChilden, ParseBlockChain(child))
	}
	producedBlock.Children = producedBlockChilden
	// fmt.Println("finshed copying the chain, the current hash is: ", producedBlock.CurrentHash)
	return producedBlock
}

func CopyBlockChainPayload(thisBlock *Block) BlockPayloadStruct {
	producedBlockPayload := BlockPayloadStruct{
		CurrentHash:       thisBlock.CurrentHash,
		PreviousHash:      thisBlock.PreviousHash,
		R:                 *thisBlock.R,
		S:                 *thisBlock.S,
		CurrentOPs:        thisBlock.CurrentOPs,
		DistanceToGenesis: thisBlock.DistanceToGenesis,
		Nonce:             thisBlock.Nonce,
		SolverPublicKey:   pubKeyToString(*thisBlock.SolverPublicKey),
	}
	var producedBlockChilden []BlockPayloadStruct
	for _, child := range thisBlock.Children {
		producedBlockChilden = append(producedBlockChilden, CopyBlockChainPayload(child))
	}
	producedBlockPayload.Children = producedBlockChilden
	// fmt.Println("finshed copying the chain, the current hash is: ", producedBlock.CurrentHash)
	return producedBlockPayload
}

func findDeepestBlocks(b *Block, depth int) (*Block, int) {
	if len(b.Children) == 0 {
		return b, depth
	} else {
		childrenDepth := make([]BlockDepthpair, 0)
		// Find depth of all children
		for _, child := range b.Children {
			block, depth := findDeepestBlocks(child, depth+1)
			childrenDepth = append(childrenDepth, BlockDepthpair{block, depth})
		}
		localMax := 0
		localMaxBlock := &Block{}

		// Pick the deepest blck from children
		for _, tuple := range childrenDepth {
			if tuple.depth.(int) > localMax {
				localMax = tuple.depth.(int)
				localMaxBlock = tuple.block.(*Block)
			}
		}
		return localMaxBlock, localMax
	}
}

func printBlock(m *Block) {
	// fmt.Println("inside printblock")
	// fmt.Println(m.Children)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if m.PreviousHash == "" {
		strconv.Itoa(len(m.Children))
		fmt.Println("Genesis BLOCK:    " + m.CurrentHash[0:5] + " has this " + strconv.Itoa(len(m.Children))  + " children")
	} else {
		fmt.Println(m.CurrentHash[0:5] + " has this " + strconv.Itoa(len(m.Children))  + " children and its parent is " + m.PreviousHash[0:5])
	}
	for _, c := range m.Children {
		printBlock(c)
	}

}


func getBlockDistanceFromGensis ( blk *Block, blkHash string) int  {

	if blk == nil {
		return -1
	}

	if blk.CurrentHash == blkHash {
		return 0
	}

	if len(blk.Children) == 0 {
		return -1
	}

	distArray := make([]int, len(blk.Children))
	for i,subB := range blk.Children {
		distArray[i] = getBlockDistanceFromGensis(subB, blkHash)
	}

	return 1 + maxArray(distArray)
}

func maxArray(array []int) int {

	if len(array) == 0 {
		return 0
	}

	maxNum := array[0]

	for _, num := range array {

		if (num > maxNum) {
			maxNum = num
		}
	}
	return maxNum
}
