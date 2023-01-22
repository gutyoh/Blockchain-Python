package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
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
	MagicNumber  int32
	PreviousHash string
	Hash         string
	Data         []string
	BuildTime    int64
	Miner        uint
}

func (b *Block) Print(nState string) {
	fmt.Printf("\nBlock:\n")
	fmt.Printf("Created by miner #%d\n", b.Miner)
	fmt.Printf("Id: %d\n", b.ID)
	fmt.Printf("Timestamp: %d\n", b.Timestamp.UnixMilli())
	fmt.Printf("Magic number: %d\n", b.MagicNumber)
	fmt.Printf("Hash of the previous block:\n%s\n", b.PreviousHash)
	fmt.Printf("Hash of the block:\n%s\n", b.Hash)

	if b.ID == 1 {
		fmt.Printf("Block data: no messages\n")
		fmt.Printf("Block was generating for %d seconds\n", b.BuildTime)
		fmt.Printf("%s\n", nState)

		b.GetBlockData()
	} else {
		fmt.Printf("Block data:\n")
		if len(b.Data) == 0 {
			fmt.Printf("no messages\n")
		} else {
			data := strings.Join(b.Data, "\n")
			if data != "" {
				fmt.Printf("%s\n", data)
				fmt.Printf("Block was generating for %d seconds\n", b.BuildTime)
				fmt.Printf("%s\n", nState)
			}

			// Delete messages from the block
			b.Data = nil
			if b.ID < 5 {
				b.GetBlockData()
			}
		}
	}

	//if b.ID > 1 {
	//	b.GetBlockData()
	//}
	//data := strings.Join(b.Data, "\n")
	//
	//if b.ID == 1 {
	//	fmt.Printf("Block data: %s\n", data)
	//	fmt.Printf("Block was generating for %d seconds\n", b.BuildTime)
	//	fmt.Printf("%s\n", nState)
	//} else {
	//	fmt.Printf("Block data:\n%s\n", data)
	//	fmt.Printf("Block was generating for %d seconds\n", b.BuildTime)
	//	fmt.Printf("%s\n", nState)
	//}
}

func (b *Block) CalculateHash() string {
	magicNumber := fmt.Sprintf("%d", b.MagicNumber)
	id := fmt.Sprintf("%d", b.ID)
	timestamp := fmt.Sprintf("%s", b.Timestamp)

	sum := sha256.Sum256([]byte(id + timestamp + b.PreviousHash + magicNumber))
	return fmt.Sprintf("%x", sum)
}

func (b *Block) GetBlockData() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter message to send to the blockchain:")
	for scanner.Scan() {
		if scanner.Text() == "" {
			break
		}
		b.Data = append(b.Data, scanner.Text())
	}
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
	Chain []*Block
}

func (bc *Blockchain) Init() {
	bc.Chain = []*Block{bc.CreateGenesisBlock()}
}

func (bc *Blockchain) CreateGenesisBlock() *Block {
	timestamp := time.Now()
	magicNumber := rand.Int31()
	miner := rand.Intn(10)
	hash := sha256.Sum256([]byte("Genesis block" + fmt.Sprintf("%d", magicNumber)))

	return &Block{ID: 1, Hash: fmt.Sprintf("%x", hash), MagicNumber: magicNumber, Miner: uint(miner),
		Timestamp: timestamp, PreviousHash: "0" /*Data: []string{"no messages"}*/}
}

func PrintGenesisBlock(difficulty int, hyperCoin *Blockchain, prefix string) (int, string) {
	difficulty++
	hyperCoin.Chain[0].Print(fmt.Sprintf(nIncreased, difficulty))
	prefix = strings.Repeat("0", difficulty)
	return difficulty, prefix
}

func main() {
	var difficulty int
	var prefix string

	hyperCoin := new(Blockchain)
	hyperCoin.Init()

	difficulty, prefix = PrintGenesisBlock(difficulty, hyperCoin, prefix)

	for i := 0; i < 4; i++ {
		next := make(chan Block)
		done := make(chan struct{})
		creator := rand.Intn(10)
		go MineBlock(hyperCoin.Chain[i], prefix, uint(creator), next, done)

		newBlock := <-next

		pendingMessages := hyperCoin.Chain[i].Data
		newBlock.Data = append(newBlock.Data, pendingMessages...)

		close(done)

		hyperCoin.Chain = append(hyperCoin.Chain, &newBlock)
		var nState string

		switch {
		case newBlock.BuildTime < 5:
			difficulty++
			nState = fmt.Sprintf(nIncreased, difficulty)
			prefix = strings.Repeat("0", difficulty)
		case newBlock.BuildTime > 10:
			difficulty--
			nState = nDecreased
			prefix = strings.Repeat("0", difficulty)
		default:
			nState = nStays
		}
		newBlock.Print(nState)
	}
}
