package CLI

import (
	"COMP5567-BlockChain/features"
	"fmt"
)

func (cli *CLI) ReindexUTXO(nodeID string) {
	blockchain := features.NewBlockChain(nodeID)
	UTXOSet := features.UTXOSet{blockchain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transaction in the UTXO set.\n", count)
}
