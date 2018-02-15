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

func monitor(minerNeighbourAddr string, miner MinerStruct, heartBeatInterval time.Duration) {
	for {
		fmt.Println("Duration is ", heartBeatInterval)

		allNeighbour.Lock()

		log.Println("MONITOR Time is: ", time.Now().UnixNano())

		log.Println("MONITOR AT %v", (time.Now().UnixNano() - allNeighbour.all[minerNeighbourAddr].RecentHeartbeat))
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

// func copyBlock(thisBlock *Block) *Block {
//
// 	producedBlock := &Block{
// 		CurrentHash:       thisBlock.CurrentHash,
// 		PreviousHash:      thisBlock.PreviousHash,
// 		UserSignature:     thisBlock.UserSignature,
// 		CurrentOP:         thisBlock.CurrentOP,
// 		Children:          make([]*Block, 0),
// 		DistanceToGenesis: thisBlock.DistanceToGenesis,
// 	}
// 	return producedBlock
// }

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	// fmt.Println(str)
	return str
}

func doProofOfWork(m *MinerStruct, nonce string, numberOfZeroes int, delay int, newOPs []Operation, leadingBlock *Block) *Block {
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
			LeafNodesMap.Lock()
			delete(LeafNodesMap.all, leadingBlock.CurrentHash)
			LeafNodesMap.all[recievedBlock.CurrentHash] = recievedBlock
			LeafNodesMap.Unlock()
			return recievedBlock
		case opFromMineNode := <-m.RecievedOpSig:
			fmt.Println("A miner sent me an operation from its art node")
			m.OPBuffer = append(m.OPBuffer, opFromMineNode)
			newOPs = m.OPBuffer
			fmt.Println("M-UPDATED OPERATION LIST FROM MINERS " + AllOperationsCommands(m.OPBuffer))
			nonce = leadingBlock.CurrentHash + opFromMineNode.Command + pubKeyToString(m.PairKey.PublicKey)
		case opFromArtnode := <-m.RecievedArtNodeSig:
			fmt.Println("ArtNode asked for this operation")
			nonce = leadingBlock.CurrentHash + opFromArtnode.Command + pubKeyToString(m.PairKey.PublicKey)
			fmt.Println(leadingBlock.CurrentHash + opFromArtnode.Command + pubKeyToString(m.PairKey.PublicKey))
			visitedMiners := []*MinerStruct{m}
			m.OPBuffer = append(m.OPBuffer, opFromArtnode)
			newOPs = m.OPBuffer
			fmt.Println("A-UPDATED OPERATION LIST FROM ART NODE" + AllOperationsCommands(m.OPBuffer))
			m.FloodOperation(&opFromArtnode, &visitedMiners)
		default:
			guessString := strconv.FormatInt(i, 10)

			hash := computeNonceSecretHash(nonce, guessString)
			if hash[32-numberOfZeroes:] == zeroes {
				log.Println("Found the hash, it is: ", hash)
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
	fmt.Println("I'm receiving the blockchain")
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
	fmt.Println("start copying the chain")
	fmt.Println(thisBlock.CurrentHash)
	fmt.Printf("%+v", thisBlock.SolverPublicKey)
	producedBlockPayload := BlockPayloadStruct{
		CurrentHash:       thisBlock.CurrentHash,
		PreviousHash:      thisBlock.PreviousHash,
		R:                 *thisBlock.R,
		S:                 *thisBlock.S,
		CurrentOPs:         thisBlock.CurrentOPs,
		DistanceToGenesis: thisBlock.DistanceToGenesis,
		Nonce:             thisBlock.Nonce,
		SolverPublicKey:   pubKeyToString(*thisBlock.SolverPublicKey),
	}
	fmt.Println("I got here")
	var producedBlockChilden []BlockPayloadStruct
	for _, child := range thisBlock.Children {
		producedBlockChilden = append(producedBlockChilden, CopyBlockChainPayload(child))
	}
	producedBlockPayload.Children = producedBlockChilden
	// fmt.Println("finshed copying the chain, the current hash is: ", producedBlock.CurrentHash)
	return producedBlockPayload
}

func deepestBlock(m *Block) *Block {
	if len(m.Children) == 0 {
		return m
	}
	return deepestBlock(m.Children[0])
}

func printBlock(m *Block) {
	// fmt.Println("inside printblock")
	// fmt.Println(m.Children)
	fmt.Printf("%v -> ", len(m.Children))
	for _, c := range m.Children {
		printBlock(c)
	}
}
