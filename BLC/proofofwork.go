package BLC

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var maxNonce = math.MaxInt64
var targetBits = 5

type ProofOfWork struct {
	block  *Block   //当前要验证的区块
	target *big.Int //大的数字，挖矿的难度
}

func NewProofOfWoek(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{block: block, target: target}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{pow.block.PrevBlockHash,
		pow.block.HashTransactions(),
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce)),
	}, []byte{})
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 1
	fmt.Println("正在挖矿...")

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		//fmt.Println(string(data))
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)


		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target)  == -1 {
			break
		}else {
			nonce ++
		}
	}
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool  {
	var hashInt big.Int
	data :=pow.prepareData(pow.block.Nonce)
	hash :=sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
