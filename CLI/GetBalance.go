package CLI

import (
	"COMP5567-BlockChain/features"
	"COMP5567-BlockChain/utils"
	"fmt"
	"log"
)

func (cli *CLI) GetBalance(address, nodeID string) {
	if !features.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	blockchain := features.NewBlockChain(nodeID)
	UTXOSet := features.UTXOSet{blockchain}
	defer blockchain.GetDB().Close()

	balance := 0
	pubKeyHash := utils.Base64Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
