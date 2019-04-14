package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

type Blockchain struct {
	Tip []byte  //最后区块的哈希
	DB *bolt.DB
}

type BlockchainIterator struct {
	CurrentHash []byte  //当前正在迭代区块的哈希
	DB *bolt.DB
}
const DBFile = "D:/Gopath/src/chain/publicChain/21-server/block_chain.db"
const BlocksBucket  = "blocks"
const genesisCoinbaseData = "我是一个创世区块"
/**
挖矿，新增区块到区块链中
 */
func (bc *Blockchain) MineBlock(transaction []*Transaction) *Block  {
	var lastHash []byte

	for _, tx := range transaction {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	//获取区块最后一块哈希
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	nowBlock := NewBlock(transaction,bc.Tip)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		err := b.Put(nowBlock.Hash,nowBlock.Serialize())
		if err == nil {
			err = b.Put([]byte("l"),nowBlock.Hash)
			if err == nil {
				bc.Tip = nowBlock.Hash
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return nowBlock
}



// 用创世区块新建一个区块链
func NewBlockchain(address string) *Blockchain {

	var tip []byte
	db, err := bolt.Open(DBFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

/**
创建创世区块
 */
func CreateBlockchain(address string)  *Blockchain{
	var tip []byte
	db, err := bolt.Open(DBFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		if b == nil {
			cbtx := NewCoinbaseTX(address,genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)
			b,err = tx.CreateBucket([]byte(BlocksBucket))
			if err == nil {
				err = b.Put(genesis.Hash,genesis.Serialize())
				if err != nil {
					log.Panic(err)
				} else {
					err = b.Put([]byte("l"),genesis.Hash)
					if err == nil {
						tip = genesis.Hash
					}
				}
			}
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	return &Blockchain{tip,db}
}

func (bc *Blockchain)Iterator() *BlockchainIterator  {
	return &BlockchainIterator{bc.Tip,bc.DB}
}

func (bci *BlockchainIterator)Next() *Block{
	var block *Block
	var preHash []byte
	err := bci.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		if b!=nil {
			block = DeserializeBlock(b.Get(bci.CurrentHash))
			preHash = block.PrevBlockHash;
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bci.CurrentHash = preHash
	return block
	//return &BlockchainIterator{preHash,bci.DB}
}

/**
找到未消费的输出
1，编列整个区块链中的区块，找到每个区块中的交易
2，编列每个区块中的交易
3，如果交易ID恰好在spentTXOs中的值不为空，则说明改交易中的某些输出已经被消费，则要找出其中为消费的部分
4，编列每个输出，如果输出的编号在spentTXOs的当前交易的值之中，则代表这输出已经被消费
 */
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		for _, tx := range block.Transcations {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							fmt.Println("111")
							continue Outputs
						}
					}
				}
				fmt.Println("222")
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			fmt.Println("333")
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

//查找所有未使用的交易输出，并返回删除了已使用的输出的交易
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transcations {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte,amount int) (int,map[string][]int) {
	unspentOutputs :=make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated :=0

	Work:
		for _,tx :=range unspentTXs{
			txID := hex.EncodeToString(tx.ID)
			for outIdx,out := range tx.Vout{
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount{
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID],outIdx)

					if accumulated >= amount {
						break Work
					}
				}
			}

		}
	return accumulated, unspentOutputs
}

/**
编列区块链，找到对应ID的交易
 */
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transcations {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("没有找到对应的ID的交易")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, _ := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, _ := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

// GetBestHeight 获取最后一块区块的高度
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// AddBlock 将区块保持到blockchain中
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.Tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// GetBlock 根据哈希值返回对应的区块
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes 获取区块链中所有哈希的区块的列表
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}