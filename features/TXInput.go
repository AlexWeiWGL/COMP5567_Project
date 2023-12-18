package features

import "bytes"

type TXInput struct {
	TXid      []byte
	Value     int
	Signature []byte
	PublicKey []byte
}

func (in *TXInput) isValidTX(publicKey []byte) bool {
	thisHash := HashPubKey(in.PublicKey)

	return bytes.Compare(thisHash, publicKey) == 0
}
