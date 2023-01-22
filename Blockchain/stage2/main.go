package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

type Block struct {
	ID           int
	Timestamp    time.Time
	PreviousHash string
	Hash         string
	MagicNumber  int64
	BuildTime    int64
	Nonce        int
}

func (b *Block) Init(timestamp time.Time, previousHash string) {
	b.ID = 1
	b.Timestamp = timestamp
	b.PreviousHash = previousHash
	b.Hash = b.CalculateHash()
	b.Nonce = 0
}

func (b *Block) CalculateHash() string {
	nonce := fmt.Sprintf("%d", b.Nonce)
	magicNumber := fmt.Sprintf("%d", b.MagicNumber)
	sum := sha256.Sum256([]byte(b.PreviousHash + b.Timestamp.String() + nonce + magicNumber))
	return fmt.Sprintf("%x", sum)
}

func (b *Block) MineBlock(difficulty int) {
	if difficulty < 1 {
		b.Hash = b.CalculateHash()
	}

	for substr(b.Hash, 0, difficulty) != strings.Repeat("0", difficulty) {
		b.Nonce++
		b.Hash = b.CalculateHash()
	}
}

func (b *Block) Print() {
	fmt.Printf("\nBlock:\n"+
		"Id: %d\n"+
		"Timestamp: %d\n"+
		"Magic number: %d\n"+
		"Hash of the previous block:\n%s\n"+
		"Hash of the block:\n%s\n"+
		"Block was generating for %d seconds\n",
		b.ID, b.Timestamp.Nanosecond(), b.MagicNumber, b.PreviousHash, b.Hash, b.BuildTime)
}

type Blockchain struct {
	Chain      []Block
	Difficulty int
}

func (bc *Blockchain) Init(difficulty int) {
	bc.Chain = []Block{bc.CreateGenesisBlock()}
	bc.Difficulty = difficulty
}

func (bc *Blockchain) CreateGenesisBlock() Block {
	timestamp := time.Now()
	magicNumber := rand.NewSource(timestamp.UnixNano())
	hash := sha256.Sum256([]byte("Genesis block" + fmt.Sprintf("%d", magicNumber)))

	return Block{ID: 1, Timestamp: time.Now(), PreviousHash: "0",
		Hash: fmt.Sprintf("%x", hash), MagicNumber: magicNumber.Int63()}
}

func main() {
	fmt.Print("Enter how many zeros the hash must start with: ")
	var difficulty int
	fmt.Scan(&difficulty)

	hyperChain := new(Blockchain)
	hyperChain.Init(difficulty)

	for i := 1; i < 5; i++ {
		block := new(Block)
		start := time.Now()
		block.ID = i + 1
		block.Timestamp = time.Now()
		block.MagicNumber = rand.NewSource(block.Timestamp.UnixNano()).Int63()
		block.PreviousHash = hyperChain.Chain[len(hyperChain.Chain)-1].Hash
		block.MineBlock(hyperChain.Difficulty)
		end := time.Now()
		block.BuildTime = end.Unix() - start.Unix()
		hyperChain.Chain = append(hyperChain.Chain, *block)
	}

	for _, block := range hyperChain.Chain {
		block.Print()
	}
}
