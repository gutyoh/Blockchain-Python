package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
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
	ID                  uint
	Timestamp           time.Time
	MagicNumber         int32
	PreviousHash        string
	Hash                string
	Transactions        []Transaction
	PendingTransactions []Transaction
	BuildTime           int64
	Miner               uint
}

func (b *Block) Print(bc *Blockchain, nState string) {
	fmt.Printf("\nBlock:\n")
	fmt.Printf("Created by miner #%d\n", b.Miner)
	fmt.Printf("Id: %d\n", b.ID)
	fmt.Printf("Timestamp: %d\n", b.Timestamp.UnixMilli())
	fmt.Printf("Magic number: %d\n", b.MagicNumber)
	fmt.Printf("Hash of the previous block:\n%s\n", b.PreviousHash)
	fmt.Printf("Hash of the block:\n%s\n", b.Hash)

	if b.ID == 1 {
		fmt.Printf("Block data: No transactions\n")
		fmt.Printf("Block was generating for %d seconds\n", b.BuildTime)
		fmt.Printf("%s\n", nState)

		// After printing the first block, ask the user if it wants to add transactions:
		b.GetTransactionData(bc)
	} else {
		fmt.Printf("Block data:\n")
		if len(b.Transactions) == 0 {
			fmt.Printf("No transactions\n")
		} else {
			for i, transaction := range b.Transactions {
				fmt.Printf("Transaction #%d:\n", i+1)
				fmt.Printf("From: %s | To: %s | Amount: %d VC\n",
					transaction.FromAddress, transaction.ToAddress, transaction.Amount)
				// fmt.Printf("Signature: %s\n", b.SignTransaction(transaction))
				fmt.Printf("Signature: %s\n", transaction.Signature)
				// fmt.Printf("Public key: %s\n", base64.StdEncoding.EncodeToString(bytes))
				fmt.Printf("Public key: %s\n", transaction.PublicKey)
			}
			fmt.Printf("Block was generating for %d seconds\n", b.BuildTime)
			fmt.Printf("%s\n", nState)

			// Ask the user to add additional transactions
			b.GetTransactionData(bc)
		}
	}
}

func (b *Block) CalculateHash() string {
	magicNumber := fmt.Sprintf("%d", b.MagicNumber)
	id := fmt.Sprintf("%d", b.ID)
	timestamp := fmt.Sprintf("%s", b.Timestamp)

	sum := sha256.Sum256([]byte(id + timestamp + b.PreviousHash + magicNumber))
	return fmt.Sprintf("%x", sum)
}

type Transaction struct {
	FromAddress string
	FromUser    string

	ToAddress string
	ToUser    string

	Amount    int
	Signature string
	PublicKey string
}

func (b *Block) GetTransactionData(bc *Blockchain) {
	fmt.Printf("\nEnter how many transactions you want to perform:\n")
	var nTx int
	fmt.Scanln(&nTx)

	for i := 0; i < nTx; i++ {
		// Ask the user the from username
		fmt.Printf("From username:\n")
		var fromAddress string
		fmt.Scanln(&fromAddress)

		// Ask the user the to address
		fmt.Printf("To username:\n")
		var toAddress string
		fmt.Scanln(&toAddress)

		// Ask the user the amount
		fmt.Printf("VC Amount:\n")
		var amount int
		fmt.Scanln(&amount)

		// Check if the user has enough VC in his wallet
		if bc.GetWalletBalance(fromAddress) < amount {
			fmt.Println("Transaction is not valid — not enough VC in wallet")
			fmt.Println("From:", fromAddress, "To:", toAddress, "Amount:", amount)
			fmt.Println(fromAddress, "current balance:", bc.GetWalletBalance(fromAddress))
			continue
		} else {
			fmt.Println("Transaction is valid")
			currentBalance := bc.GetWalletBalance(fromAddress)
			fmt.Println(fromAddress, "current balance:", currentBalance)
			fmt.Println("From:", fromAddress, "To:", toAddress, "Amount:", amount, "VC")
			fmt.Println(fromAddress, "remaining balance:", currentBalance-amount)
			fmt.Println("----------------------------------------")
			fmt.Println(toAddress, "new balance:", bc.GetWalletBalance(toAddress)+amount)
			fmt.Println()

			// Sign the transaction
			signature := b.SignTransaction(Transaction{FromAddress: fromAddress, ToAddress: toAddress, Amount: amount})

			// Get the public key
			publicKey := b.GetPrivateKey().PublicKey
			bytes, err := x509.MarshalPKIXPublicKey(&publicKey)
			if err != nil {
				log.Fatal(err)
			}
			publicKeyString := base64.StdEncoding.EncodeToString(bytes)

			b.Transactions = append(b.Transactions, Transaction{
				FromAddress: fromAddress,
				ToAddress:   toAddress,
				Amount:      amount,
				Signature:   signature,
				PublicKey:   publicKeyString,
			})
		}
	}
}

func (b *Block) SignTransaction(transaction Transaction) string {
	hash := sha256.Sum256([]byte(transaction.FromAddress +
		transaction.ToAddress + fmt.Sprintf("%d", transaction.Amount)))

	bytes, err := ecdsa.SignASN1(cryptoRand.Reader, b.GetPrivateKey(), hash[:])
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(bytes)
}

func (b *Block) GetPrivateKey() *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptoRand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	return privateKey
}

func (b *Block) GenerateMessageID(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func (b *Block) AddTransactions(transactions []Transaction) {
	b.Transactions = append(b.Transactions, transactions...)
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

	// Sign the reward transaction
	signature := b.SignTransaction(Transaction{FromAddress: "Blockchain",
		ToAddress: fmt.Sprintf("miner #%d", creator), Amount: 100})

	// Get the public key
	publicKey := b.GetPrivateKey().PublicKey
	bytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		log.Fatal(err)
	}
	publicKeyString := base64.StdEncoding.EncodeToString(bytes)

	// Create a transaction that rewards 100 VC to the miner that mined the block
	b.Transactions = append(b.Transactions, Transaction{
		FromAddress: "Blockchain",
		ToAddress:   fmt.Sprintf("miner #%d", creator),
		Amount:      100,
		Signature:   signature,
		PublicKey:   publicKeyString,
	})

	next <- b
}

type Blockchain struct {
	Chain []*Block
}

func (bc *Blockchain) Init() {
	bc.Chain = []*Block{bc.CreateGenesisBlock()}
}

func (bc *Blockchain) GetWalletBalance(address string) int {
	balance := 100

	for _, block := range bc.Chain {
		//if len(bc.Chain) > 1 && address == block.Transactions[i].FromAddress {
		//	balance += 100
		//}

		for _, tx := range block.Transactions {
			if address == tx.FromAddress {
				balance -= tx.Amount
			}

			if address == tx.ToAddress {
				balance += tx.Amount
			}
		}
	}
	return balance
}

func (bc *Blockchain) CreateGenesisBlock() *Block {
	timestamp := time.Now()
	magicNumber := rand.Int31()
	miner := rand.Intn(10)
	hash := sha256.Sum256([]byte("Genesis block" + fmt.Sprintf("%d", magicNumber)))

	return &Block{ID: 1, Hash: fmt.Sprintf("%x", hash), MagicNumber: magicNumber, Miner: uint(miner),
		Timestamp: timestamp, PreviousHash: "0", Transactions: []Transaction{}}
}

func PrintGenesisBlock(difficulty int, hyperCoin *Blockchain, prefix string) (int, string) {
	difficulty++
	hyperCoin.Chain[0].Print(hyperCoin, fmt.Sprintf(nIncreased, difficulty))
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
		pendingTransactions := hyperCoin.Chain[i].Transactions
		newBlock.AddTransactions(pendingTransactions)
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
		newBlock.Print(hyperCoin, nState)
	}
}
