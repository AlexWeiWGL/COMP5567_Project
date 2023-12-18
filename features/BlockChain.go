package features

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Genesis Coinbase..."

type BlockChain struct {
	Tip []byte
	DB  *bolt.DB
}

func (blockchain *BlockChain) GetDB() *bolt.DB {
	return blockchain.DB
}

func CreateBlockChain(address, nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		fmt.Println("BlockChain exists!")
		os.Exit(1)
	}

	var tip []byte
	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(transaction *bolt.Tx) error {
		b, err := transaction.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.GetHash(), genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.GetHash()

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{tip, db}
	return &blockchain
}

func NewBlockChain(nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No BlockChain found !!! Need to create...")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bc := BlockChain{tip, db}
	return &bc
}

func (blockchain *BlockChain) AddBlock(block *Block) {
	err := blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDB := b.Get(block.GetHash())

		if blockInDB != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.GetHash(), blockData)
		if err != nil {
			log.Fatal(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.GetHeight() > lastBlock.GetHeight() {
			err = b.Put([]byte("l"), block.GetHash())
			if err != nil {
				log.Fatal(err)
			}
			blockchain.Tip = block.GetHash()
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func (blockchain *BlockChain) FindTransaction(Id []byte) (Transaction, error) {
	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, Id) == 0 {
				return *tx, nil
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found !!!")
}

func (blockchain *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	iter := blockchain.Iterator()

	for {
		block := iter.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.TXOutputs {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCionBase() == false {
				for _, in := range tx.TXInputs {
					inTxID := hex.EncodeToString(in.TXid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Value)
				}
			}
		}
		if len(block.PreviousHash) == 0 {
			break
		}
	}
	return UTXO
}

func (blockchain *BlockChain) Iterator() *BlockChainIterator {
	iterator := &BlockChainIterator{blockchain.Tip, blockchain.DB}

	return iterator
}

func (blockchain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Height
}

func (blockchain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found !!!")
		}

		block = *DeserializeBlock(blockData)
		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

func (bc *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	iter := bc.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PreviousHash) == 0 {
			break
		}
	}
	return blocks
}

func (bc *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)
		lastHeight = block.Height
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	nBlock := NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(nBlock.Hash, nBlock.Serialize())
		if err != nil {
			log.Fatal(err)
		}

		err = b.Put([]byte("l"), nBlock.Hash)
		if err != nil {
			log.Fatal(err)
		}

		bc.Tip = nBlock.Hash
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return nBlock
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.TXInputs {
		prevTX, err := bc.FindTransaction(in.TXid)
		if err != nil {
			log.Fatal(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCionBase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.TXInputs {
		prevTX, err := bc.FindTransaction(in.TXid)
		if err != nil {
			log.Fatal(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
