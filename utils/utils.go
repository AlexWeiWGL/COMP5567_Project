package utils

import (
	"bytes"
	"encoding/binary"
	"log"
)

func Int2Hex(number int64) []byte {
	inputBuff := new(bytes.Buffer)
	err := binary.Write(inputBuff, binary.BigEndian, number)
	if err != nil {
		log.Panic("Binary write Failed !!!", err)
	}

	return inputBuff.Bytes()
}
