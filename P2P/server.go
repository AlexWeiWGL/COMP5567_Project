package P2P

import (
	"COMP5567-BlockChain/features"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]features.Transaction)

type Address struct {
	AddressList []string
}

type BlockSender struct {
	AddressFrom string
	Block       []byte
}

type BlockSenderAddr struct {
	AddressFrom string
}

type Data struct {
	AddressFrom string
	Type        string
	ID          []byte
}

type Inv struct {
	AddressFrom string
	Type        string
	Items       [][]byte
}

type TX struct {
	AddressFrom string
	Transaction []byte
}

type version struct {
	Version     int
	BestHeight  int
	AddressFrom string
}

func Command2Bytes(command string) []byte {
	var bytes [commandLength]byte
	for i, j := range command {
		bytes[i] = byte(j)
	}
	return bytes[:]
}

func Bytes2Command(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func ExtractCommand(request []byte) []byte {
	return request[:commandLength]
}

func RequestBlock() {
	for _, node := range knownNodes {
		SendGetBlock(node)
	}
}

func GetKnownNodes() []string {
	return knownNodes
}

func SendAddress(address string) {
	nodes := Address{knownNodes}
	nodes.AddressList = append(nodes.AddressList, nodeAddress)
	payload := GobEncode(nodes)
	request := append(Command2Bytes("Address"), payload...)

	SendData(address, request)
}

func SendBlock(address string, b *features.Block) {
	data := BlockSender{nodeAddress, b.Serialize()}
	payload := GobEncode(data)
	request := append(Command2Bytes("block"), payload...)

	SendData(address, request)
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	encryption := gob.NewEncoder(&buff)
	err := encryption.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func SendData(address string, data []byte) {
	connection, err := net.Dial(protocol, address)
	if err != nil {
		fmt.Printf("%s is not available\n", address)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != address {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes
		return
	}
	defer connection.Close()
	_, err = io.Copy(connection, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func SendInv(address, kind string, items [][]byte) {
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(Command2Bytes("Inv"), payload...)

	SendData(address, request)
}

func SendGetData(address, kind string, id []byte) {
	payload := GobEncode(Data{nodeAddress, kind, id})
	request := append(Command2Bytes("Data"), payload...)

	SendData(address, request)
}

func SendTX(address string, transactions *features.Transaction) {
	data := TX{nodeAddress, transactions.Serialize()}
	payload := GobEncode(data)
	request := append(Command2Bytes("TX"), payload...)

	SendData(address, request)
}

func SendVersion(address string, blockchain *features.BlockChain) {
	bestHeight := blockchain.GetBestHeight()
	payload := GobEncode(version{nodeVersion, bestHeight, nodeAddress})

	request := append(Command2Bytes("version"), payload...)

	SendData(address, request)
}

func HandleAddress(request []byte) {
	var buff bytes.Buffer
	var payload Address

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddressList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	RequestBlock()
}

func HandleBlock(request []byte, blockchain *features.BlockChain) {
	var buff bytes.Buffer
	var payload BlockSender

	buff.Write(request[commandLength:])
	dec := gob.NewEncoder(&buff)
	err := dec.Encode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := features.DeserializeBlock(blockData)

	fmt.Printf("Received a new Block !!!")
	blockchain.AddBlock(block)

	fmt.Printf("Added block %x\n", block.GetHash())

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddressFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := features.UTXOSet{blockchain}
		UTXOSet.Reindex()
	}
}

func HandleGetData(request []byte, blockchain *features.BlockChain) {
	var buff bytes.Buffer
	var payload Data

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := blockchain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		SendBlock(payload.AddressFrom, &block)
	}

	if payload.Type == "TX" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTX(payload.AddressFrom, &tx)
	}
}

func HandleInv(request []byte, blockchain *features.BlockChain) {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		SendGetData(payload.AddressFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "TX" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddressFrom, "TX", txID)
		}
	}
}

func HandleGetBlocks(request []byte, blockchain *features.BlockChain) {
	var buff bytes.Buffer
	var payload Data

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := blockchain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		SendBlock(payload.AddressFrom, &block)
	}

	if payload.Type == "TX" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTX(payload.AddressFrom, &tx)
	}
}

func HandleTX(request []byte, blockchain *features.BlockChain) {
	var buff bytes.Buffer
	var payload TX

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := features.DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddressFrom {
				SendInv(node, "TX", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*features.Transaction

			for id := range mempool {
				tx := mempool[id]
				if blockchain.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := features.NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := blockchain.MineBlock(txs)
			UTXOSet := features.UTXOSet{blockchain}
			UTXOSet.Reindex()

			fmt.Println("New block is mined!!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					SendInv(node, "block", [][]byte{newBlock.GetHash()})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func HandleVersion(request []byte, blockchain *features.BlockChain) {
	var buff bytes.Buffer
	var payload version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := blockchain.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		SendGetBlock(payload.AddressFrom)
	} else if myBestHeight > foreignerBestHeight {
		SendVersion(payload.AddressFrom, blockchain)
	}

	if !nodeIsKnown(payload.AddressFrom) {
		knownNodes = append(knownNodes, payload.AddressFrom)
	}
}

func HandleConnection(conn net.Conn, blockchain *features.BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := Bytes2Command(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "Address":
		HandleAddress(request)
	case "block":
		HandleBlock(request, blockchain)
	case "Inv":
		HandleInv(request, blockchain)
	case "BlockSenderAddr":
		HandleGetBlocks(request, blockchain)
	case "Data":
		HandleGetData(request, blockchain)
	case "TX":
		HandleTX(request, blockchain)
	case "version":
		HandleVersion(request, blockchain)
	default:
		fmt.Println("Unknown Command!!!")
	}

	err = conn.Close()
	if err != nil {
		return
	}
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	blockchain := features.NewBlockChain(nodeID)

	if nodeAddress != knownNodes[0] {
		SendVersion(knownNodes[0], blockchain)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, blockchain)
	}
}

func nodeIsKnown(address string) bool {
	for _, node := range knownNodes {
		if node == address {
			return true
		}
	}
	return false
}

func SendGetBlock(address string) {
	payload := GobEncode(BlockSenderAddr{nodeAddress})
	request := append(Command2Bytes("BlockSenderAddr"), payload...)

	SendData(address, request)
}
