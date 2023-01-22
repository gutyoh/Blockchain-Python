package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	nIncreased = "N was increased to %d"
	nDecreased = "N was decreased by 1"
	nStays     = "N stays the same"
)

type Block struct {
	ID           uint
	Timestamp    time.Time
	PreviousHash string
	Hash         string
	MagicNumber  int32
	BuildTime    int64
	Miner        uint
}

func (b *Block) Print(nState string) {
	fmt.Printf("\nBlock:\n"+
		"Created by miner #%d\n"+
		"Id: %d\n"+
		"Timestamp: %d\n"+
		"Magic number: %d\n"+
		"Hash of the previous block:\n%s\n"+
		"Hash of the block:\n%s\n"+
		"Block was generating for %d seconds\n"+
		"%s\n",
		b.Miner, b.ID, b.Timestamp.UnixMilli(), b.MagicNumber, b.PreviousHash, b.Hash, b.BuildTime, nState)
}

func (b *Block) CalculateHash() string {
	magicNumber := fmt.Sprintf("%d", b.MagicNumber)
	id := fmt.Sprintf("%d", b.ID)
	timestamp := fmt.Sprintf("%s", b.Timestamp)

	sum := sha256.Sum256([]byte(id + timestamp + b.PreviousHash + magicNumber))
	return fmt.Sprintf("%x", sum)
}

func MineBlock(prevBlock *Block, prefix string, creator uint, next chan Block, done chan struct{}) {
	start := time.Now()
	b := Block{
		ID:           prevBlock.ID + 1,
		PreviousHash: prevBlock.Hash,
	}

Loop:
	for {
		select {
		case <-done:
			break Loop
		default:
			b.MagicNumber = rand.Int31()
			b.Hash = b.CalculateHash()
			if strings.HasPrefix(b.Hash, prefix) {
				break Loop
			}
		}
	}

	b.Timestamp = time.Now()
	b.BuildTime = int64(time.Since(start).Seconds())
	b.Miner = creator
	next <- b
}

type Blockchain struct {
	Chain      []*Block
	Difficulty int
}

func (bc *Blockchain) Init() {
	bc.Chain = []*Block{bc.CreateGenesisBlock()}
	bc.Difficulty = 0
}

func (bc *Blockchain) CreateGenesisBlock() *Block {
	timestamp := time.Now()
	magicNumber := rand.NewSource(timestamp.UnixNano())
	hash := sha256.Sum256([]byte("Genesis block" + fmt.Sprintf("%d", magicNumber)))

	return &Block{ID: 0, Hash: fmt.Sprintf("%x", hash), Timestamp: timestamp, PreviousHash: "0"}
}

func main() {
	difficulty := 0
	prefix := ""

	hyperCoin := new(Blockchain)
	hyperCoin.Init()

	for i := 0; i < 5; i++ {
		next := make(chan Block)
		done := make(chan struct{})
		for j := 1; j < 5; j++ {
			creator := rand.Intn(10)
			go MineBlock(hyperCoin.Chain[i], prefix, uint(creator), next, done)
		}

		newBlock := <-next
		close(done)

		hyperCoin.Chain = append(hyperCoin.Chain, &newBlock)
		var nState string
		switch {
		case newBlock.BuildTime < 1:
			difficulty++
			nState = fmt.Sprintf(nIncreased, difficulty)
			prefix = strings.Repeat("0", difficulty)
		case newBlock.BuildTime >= 2:
			difficulty--
			nState = nDecreased
			prefix = strings.Repeat("0", difficulty)
		default:
			nState = nStays
		}
		newBlock.Print(nState)
	}
}
