package CLI

import (
	"COMP5567-BlockChain/features"
	"fmt"
)

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := features.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
