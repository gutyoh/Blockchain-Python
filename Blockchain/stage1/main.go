package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	ID           int
	Timestamp    time.Time
	PreviousHash string
	Hash         string
}

func (b *Block) Init(timestamp time.Time, previousHash string) {
	b.ID = 1
	b.Timestamp = timestamp
	b.PreviousHash = previousHash
	b.Hash = b.CalculateHash()
}

func (b *Block) CalculateHash() string {
	sum := sha256.Sum256([]byte(b.PreviousHash + b.Timestamp.String()))
	return fmt.Sprintf("%x", sum)
}

func (b *Block) Print() {
	fmt.Printf("\nBlock:\n"+
		"Id: %d\n"+
		"Timestamp: %d\n"+
		"Hash of the previous block:\n%s\n"+
		"Hash of the block:\n%s\n",
		b.ID, b.Timestamp.Nanosecond(), b.PreviousHash, b.Hash)
}

func main() {
	hyperBlock := new(Block)
	hyperBlock.Init(time.Now(), "0")
	hyperBlock.Print()

	for i := 0; i < 4; i++ {
		hyperBlock.ID++
		hyperBlock.Timestamp = time.Now()
		hyperBlock.PreviousHash = hyperBlock.Hash
		hyperBlock.Hash = hyperBlock.CalculateHash()
		hyperBlock.Print()
	}
}
