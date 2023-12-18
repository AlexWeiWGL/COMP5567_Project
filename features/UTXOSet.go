package features

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

const UTXOBucket = "chainstate"

type UTXOSet struct {
	BlockChain *BlockChain
}

func (utxo UTXOSet) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := utxo.BlockChain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

func (utxo UTXOSet) FindUTXO(publicKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := utxo.BlockChain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

func (utxo UTXOSet) CountTransactions() int {
	db := utxo.BlockChain.DB
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return counter
}

func (utxo UTXOSet) Reindex() {
	db := utxo.BlockChain.DB
	bucketName := []byte(UTXOBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := utxo.BlockChain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

func (utxo UTXOSet) Update(block *Block) {
	db := utxo.BlockChain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOBucket))

		for _, tx := range block.Transactions {
			if tx.IsCionBase() == false {
				for _, in := range tx.TXInputs {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(in.TXid)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Value {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(in.TXid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(in.TXid, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}
			nOutputs := TXOutputs{}
			for _, out := range tx.TXOutputs {
				nOutputs.Outputs = append(nOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, nOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
