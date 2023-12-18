package CLI

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type CLI struct {
}

func (cli *CLI) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("	createBlockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("	createWallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("	getBalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("	listAddress - Lists all addresses from the wallet file")
	fmt.Println("	printChain - Print all the blocks of the blockchain")
	fmt.Println("	send -from ADDRESS -to ADDRESS -record RECORD -mine - Send a Record from address A to address B, if -mine is set, mine on the same node.")
	fmt.Println("	reindexUTXO - Rebuilds the UTXO set")
	fmt.Println("	startNode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
	fmt.Println(" 	switchUser -target Number - Switch the user to target")
}

type Config struct {
	nodeID int `yaml:"nodeID"`
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Panic("ERROR When READING YAML file.")
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	nodeID := config.nodeID

	nodeIDString := fmt.Sprintf("%d", nodeID)
	fmt.Println(yamlFile)
	if nodeIDString == "" {
		fmt.Printf("NODE_ID env. var is not set!!")
		os.Exit(1)
	}

	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	listAddressCmd := flag.NewFlagSet("listAddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexUTXO", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startNode", flag.ExitOnError)
	switchNodeCmd := flag.NewFlagSet("switchNode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	//switchNode := switchNodeCmd.String("target", "", "Switch the user to target")

	switch os.Args[1] {
	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createBlockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listAddresses":
		err := listAddressCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "reindexUTXO":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "startNode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "switch":
		err := switchNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.PrintUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.GetBalance(*getBalanceAddress, nodeIDString)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeIDString)
	}

	if createBlockchainCmd.Parsed() {
		cli.createBlockchain(*createBlockchainAddress, nodeIDString)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain(nodeIDString)
	}

	if listAddressCmd.Parsed() {
		cli.listAddresses(nodeIDString)
	}

	if reindexUTXOCmd.Parsed() {
		cli.ReindexUTXO(nodeIDString)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.Send(*sendFrom, *sendTo, *sendAmount, nodeIDString, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeIDstr := fmt.Sprintf("%d", nodeID)
		if nodeIDstr == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.StartNode(nodeIDString, *startNodeMiner)
	}

	//if switchNodeCmd.Parsed() {
	//	if *switchNode == "" {
	//
	//	}
	//	cli.SwitchNode(, *switchNode)
	//}

}
