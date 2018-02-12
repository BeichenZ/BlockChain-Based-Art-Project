package shared

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
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

func copyBlock(thisBlock *Block) *Block {

	producedBlock := &Block{CurrentHash: thisBlock.CurrentHash,
		PreviousHash:      thisBlock.PreviousHash,
		CurrentOP:         thisBlock.CurrentOP,
		Children:          make([]*Block, 0),
		DistanceToGenesis: thisBlock.DistanceToGenesis}
	return producedBlock
}

func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	// fmt.Println(str)
	return str
}

func doProofOfWork(m *MinerStruct, nonce string, numberOfZeroes int, delay int, newOP Operation, leadingBlock *Block) *Block {
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
			delete(m.LeafNodesMap, leadingBlock.CurrentHash)
			m.LeafNodesMap[recievedBlock.CurrentHash] = recievedBlock
			return recievedBlock
		case opFromMineNode := <-m.RecievedOpSig:
			fmt.Println("A miner sent me an operation from its art node")
			m.OPBuffer = append(m.OPBuffer, opFromMineNode)
			fmt.Println("M-UPDATED OPERATION LIST FROM MINERS " + AllOperationsCommands(m.OPBuffer))
			nonce = opFromMineNode.Command + pubKeyToString(m.PairKey.PublicKey) + leadingBlock.CurrentHash
 		case opFromArtnode := <-m.RecievedArtNodeSig:
			fmt.Println("ArtNode asked for this operation")
			nonce = opFromArtnode.Command + pubKeyToString(m.PairKey.PublicKey) + leadingBlock.CurrentHash
			visitedMiners := []*MinerStruct{m}
			m.OPBuffer = append(m.OPBuffer, opFromArtnode)
			fmt.Println("A-UPDATED OPERATION LIST FROM ART NODE" + AllOperationsCommands(m.OPBuffer))
			m.FloodOperation(&opFromArtnode, &visitedMiners)
		default:
			guessString := strconv.FormatInt(i, 10)

			hash := computeNonceSecretHash(nonce, guessString)
			if hash[32-numberOfZeroes:] == zeroes {
				log.Println("Found the hash")
				m.FoundHash = true
				return m.produceBlock(hash, newOP, leadingBlock)
			}
			i++
		}
	}
}

func pubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}

func printBlock(m *Block) {
	fmt.Printf("%v -> ", len(m.Children))
	for _, c := range m.Children {
		printBlock(c)
	}
}
