package BLC

import (
	"bytes"
	"encoding/binary"
	"log"
)

func IntToHex(i int64) []byte  {
	buf := new(bytes.Buffer)
	err := binary.Write(buf,binary.BigEndian,i)

	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

// 反转字节数组
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}