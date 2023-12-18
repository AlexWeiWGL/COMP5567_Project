package utils

import (
	"encoding/base64"
)

func Base64Encode(input []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(input))
}

func Base64Decode(input []byte) []byte {
	encode := string(input)
	decode, _ := base64.StdEncoding.DecodeString(encode)
	return decode
}
