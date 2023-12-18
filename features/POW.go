package features

import (
	"COMP5567-BlockChain/utils"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var maxNonce = math.MaxInt64

const targetBits = 16

type POW struct {
	block  *Block
	target *big.Int
}

func NewPOW(block *Block) *POW {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &POW{block, target}
	return pow
}

func (pow *POW) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PreviousHash,
			pow.block.HashTransactions(),
			utils.Int2Hex(pow.block.TimeStamp),
			utils.Int2Hex(int64(targetBits)),
			utils.Int2Hex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *POW) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Println("Mining a new Block...")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		if math.Remainder(float64(nonce), 100000) == 0 {
			fmt.Printf("\r%x", hash)
		}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:]
}

func (pow *POW) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
