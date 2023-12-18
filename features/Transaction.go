package features

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy = 10

type Transaction struct {
	ID        []byte     `json:"ID"`
	TXInputs  []TXInput  `json:"TXInputs"`
	TXOutputs []TXOutput `json:"TXOutputs"`
}

// check whether this transaction is coinbase
func (transation Transaction) IsCionBase() bool {
	return len(transation.TXInputs) == 1 && len(transation.TXInputs[0].TXid) == 0 && transation.TXInputs[0].Value == -1
}

func (transaction Transaction) Serialize() []byte {
	var encode bytes.Buffer

	encryption := gob.NewEncoder(&encode)
	err := encryption.Encode(transaction)
	if err != nil {
		log.Panic("Transaction Serialization Failed!!!", err)
	}
	return encode.Bytes()
}

func (transaction *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *transaction
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (transaction *Transaction) ModifiedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range transaction.TXInputs {
		inputs = append(inputs, TXInput{in.TXid, in.Value, nil, nil})
	}

	for _, out := range transaction.TXOutputs {
		outputs = append(outputs, TXOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{transaction.ID, inputs, outputs}
	return txCopy
}

func (transaction *Transaction) Sign(privateKey ecdsa.PrivateKey, previousTXs map[string]Transaction) {
	if transaction.IsCionBase() {
		return
	}

	for _, in := range transaction.TXInputs {
		if previousTXs[hex.EncodeToString(in.TXid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct!!!")
		}
	}

	txCopy := transaction.ModifiedCopy()

	for id, in := range txCopy.TXInputs {
		previousTX := previousTXs[hex.EncodeToString(in.TXid)]
		txCopy.TXInputs[id].Signature = nil
		txCopy.TXInputs[id].PublicKey = previousTX.TXOutputs[in.Value].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		transaction.TXInputs[id].Signature = signature
		txCopy.TXInputs[id].PublicKey = nil
	}
}

func (transaction *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction %x:", transaction.ID))

	for i, input := range transaction.TXInputs {
		lines = append(lines, fmt.Sprintf("		Input 			%d:", i))
		lines = append(lines, fmt.Sprintf("		Transaction ID:	%x", input.TXid))
		lines = append(lines, fmt.Sprintf("		Value Output: 	%d", input.Value))
		lines = append(lines, fmt.Sprintf("		Signature: 		%x", input.Signature))
		lines = append(lines, fmt.Sprintf("		PubKey:			%x", input.PublicKey))
	}

	for i, output := range transaction.TXOutputs {
		lines = append(lines, fmt.Sprintf("		Output			%d:", i))
		lines = append(lines, fmt.Sprintf("		Value:			%d:", output.Value))
		lines = append(lines, fmt.Sprintf("		PublicKeyHash:	%x:", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}

// verify the signature of Transaction Inputs
func (transaction *Transaction) Verify(previousTXs map[string]Transaction) bool {
	if transaction.IsCionBase() {
		return true
	}

	for _, in := range transaction.TXInputs {
		if previousTXs[hex.EncodeToString(in.TXid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct!!!")
		}
	}

	txCopy := transaction.ModifiedCopy()
	curve := elliptic.P256()

	for id, in := range transaction.TXInputs {
		previousTX := previousTXs[hex.EncodeToString(in.TXid)]
		txCopy.TXInputs[id].Signature = nil
		txCopy.TXInputs[id].PublicKey = previousTX.TXOutputs[in.Value].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PublicKey)
		x.SetBytes(in.PublicKey[:(keyLen / 2)])
		y.SetBytes(in.PublicKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPublicKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPublicKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.TXInputs[id].PublicKey = nil
	}
	return true
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txInput := TXInput{[]byte{}, -1, nil, []byte(data)}
	txOutput := NewTXOutput(subsidy, to)
	transaction := Transaction{nil, []TXInput{txInput}, []TXOutput{*txOutput}}
	transaction.ID = transaction.Hash()

	return &transaction
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := HashPubKey(wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for id, outs := range validOutputs {
		txID, err := hex.DecodeString(id)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc < amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	transaction := Transaction{nil, inputs, outputs}
	transaction.ID = transaction.Hash()
	UTXOSet.BlockChain.SignTransaction(&transaction, *wallet.PrivateKey)

	return &transaction
}
