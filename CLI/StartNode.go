package CLI

import (
	"COMP5567-BlockChain/P2P"
	"COMP5567-BlockChain/features"
	"fmt"
	"log"
)

func (cli *CLI) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if features.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	P2P.StartServer(nodeID, minerAddress)
}
