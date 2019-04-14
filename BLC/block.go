package BLC

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	PrevBlockHash []byte
	Transcations   []*Transaction
	Hash          []byte
	Nonce         int
	Height        int
}

func NewBlock(transcations []*Transaction, preBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), preBlockHash, transcations, []byte{},0,0}
	pow:=NewProofOfWoek(block)
	nonce,hash :=pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}
func NewGenesisBlock(coinbaae *Transaction) *Block {
	block := NewBlock([]*Transaction{coinbaae}, []byte{})
	return block
}

/*func (block *Block) SetHash() {
	timeStr := strconv.FormatInt(block.Timestamp, 10)
	timestamp := []byte(timeStr)
	//fmt.Println(timeStr,timestamp)

	headers := bytes.Join([][]byte{block.PrevBlockHash, block.Transcations, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	block.Hash = hash[:]
}*/

// 将区块序列化成字节数组
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// 反序列化
func DeserializeBlock(blockBytes []byte) *Block {

	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

/**
编列区块中的交易哈希，将他们连接并重新哈希
 */
func (b *Block) HashTransactions() []byte  {
	var txHashes [][]byte
	var txHash [32]byte

	for _,tx := range b.Transcations{
		txHashes = append(txHashes,tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes,[]byte{}))
	return txHash[:]
}



func (block *Block) Dump() {
	fmt.Printf("\nTimestamp = %v \n", block.Timestamp)
	fmt.Printf("PrevBlockHash = %x \n", block.PrevBlockHash)
	fmt.Printf("Data = %s \n", block.Transcations)
	fmt.Printf("Hash = %x \n", block.Hash)
	fmt.Printf("Nonce = %d \n", block.Nonce)
}
