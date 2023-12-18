package CLI

import (
	"COMP5567-BlockChain/features"
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !features.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	blockchain := features.CreateBlockChain(address, nodeID)
	defer blockchain.GetDB().Close()

	UTXOSet := features.UTXOSet{BlockChain: blockchain}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}
