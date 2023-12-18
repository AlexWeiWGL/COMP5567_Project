package features

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type Block struct {
	TimeStamp    int64
	Transactions []*Transaction
	PreviousHash []byte
	Hash         []byte
	Nonce        int
	Height       int
}

func NewBlock(transactions []*Transaction, previousHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), transactions, previousHash, []byte{}, 0, height}
	pow := NewPOW(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (block *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, transaction := range block.Transactions {
		transactions = append(transactions, transaction.Serialize())
	}

	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Value
}

func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(&block)
	if err != nil {
		fmt.Println("Decryption ERROR: ", err)
		return nil
	}

	return result.Bytes()
}

func DeserializeBlock(input []byte) *Block {
	var block Block

	decode := gob.NewDecoder(bytes.NewReader(input))
	err := decode.Decode(&block)
	if err != nil {
		log.Println("Decryption ERROR: ", err)
	}

	return &block
}

func (block *Block) GetHash() []byte {
	return block.Hash
}

func (block *Block) GetHeight() int {
	return block.Height
}

func (block *Block) GetPreviousHash() []byte {
	return block.PreviousHash
}

func (block *Block) GetTransactions() []*Transaction {
	return block.Transactions
}
