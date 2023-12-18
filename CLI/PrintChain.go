package CLI

import (
	"COMP5567-BlockChain/features"
	"fmt"
	"strconv"
)

func (cli *CLI) PrintChain(nodeID string) {
	blockchain := features.NewBlockChain(nodeID)
	defer blockchain.GetDB().Close()

	bci := blockchain.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============== Block %x ==============", block.GetHash())
		fmt.Printf("Height: %d\n", block.GetHeight())
		fmt.Printf("Prev. block: %x\n", block.GetPreviousHash())
		pow := features.NewPOW(block)
		fmt.Printf("POW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, transaction := range block.GetTransactions() {
			fmt.Println(transaction)
		}
		fmt.Printf("\n\n")

		if len(block.GetPreviousHash()) == 0 {
			break
		}
	}
}
