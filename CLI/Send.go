package CLI

import (
	"COMP5567-BlockChain/P2P"
	"COMP5567-BlockChain/features"
	"fmt"
	"log"
)

func (cli *CLI) Send(from, to string, amount int, nodeID string, mineNow bool) {
	if !features.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !features.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	blockchain := features.NewBlockChain(nodeID)
	UTXOSet := features.UTXOSet{blockchain}
	defer blockchain.GetDB().Close()

	wallets, err := features.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	transation := features.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbtx := features.NewCoinbaseTX(from, "")
		txs := []*features.Transaction{cbtx, transation}

		newBlock := blockchain.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		P2P.SendTX(P2P.GetKnownNodes()[0], transation)
	}

	fmt.Println("Success!")
}
