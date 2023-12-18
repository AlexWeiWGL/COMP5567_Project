package features

import (
	"COMP5567-BlockChain/utils"
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base64Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// to check the output whether can be used by the public key owner
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TXOutput {
	txOutput := &TXOutput{value, nil}
	txOutput.Lock([]byte(address))

	return txOutput
}

type TXOutputs struct {
	Outputs []TXOutput
}

func (outputs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	encryption := gob.NewEncoder(&buff)
	err := encryption.Encode(outputs)
	if err != nil {
		log.Panic("Serialization Error!!!!", err)
	}

	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	if err != nil {
		log.Panic("Serialization Decoding Error!!!", err)
	}

	return outputs
}
