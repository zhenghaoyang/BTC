package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

////添加区块
//func (bc *Blockchain) AddBlock(Transactions []*Transaction) {
//	var lastHash []byte
//	err := bc.db.View(func(tx *bolt.Tx) error {
//		block := tx.Bucket([]byte(blocksBucket))
//		lastHash = block.Get([]byte("l"))
//		return nil
//	})
//	if err != nil {
//		log.Panicln(err) //view错误
//	}
//
//	newBlock := NewBlock(Transactions, lastHash)
//	err = bc.db.Update(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte(blocksBucket))
//
//		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
//		if err != nil {
//			log.Panicln(err) //Put错误
//		}
//
//		err = bucket.Put([]byte("l"), newBlock.Hash)
//		if err != nil {
//			log.Panicln(err) //Put错误
//		}
//		return nil
//	})
//
//	//更新链上的tip
//	bc.tip = newBlock.Hash
//	if err != nil {
//		log.Panicln(err) //更新错误
//	}
//}

func CreatBlockchain(address string) *Blockchain {
	//先判断数据存在
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panicln(err) //打开错误
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(blocksBucket)) //创建数据库表
		if err != nil {
			log.Panicln(err) //更新错误
		}
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		block := RootBlock(cbtx)
		err = bucket.Put(block.Hash, block.Serialize())

		if err != nil {
			log.Panicln(err) //更新错误
		}

		err = bucket.Put([]byte("l"), block.Hash)

		if err != nil {
			log.Panicln(err) //更新错误
		}

		tip = block.Hash
		return nil
	})

	if err != nil {
		log.Panicln(err) //更新错误
	}
	bc := Blockchain{tip, db}
	return &bc

}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockchain(address string) *Blockchain {

	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		log.Panicln(err) //打开错误
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket)) //获取数据库表

		if bucket == nil { //创建新数据表
			fmt.Println("No exists blockchain,Creat one first ")
			bucket, err := tx.CreateBucket([]byte(blocksBucket)) //创建数据库表
			if err != nil {
				log.Panicln(err) //更新错误
			}

			cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
			block := RootBlock(cbtx)
			err = bucket.Put(block.Hash, block.Serialize())

			if err != nil {
				log.Panicln(err) //更新错误
			}

			err = bucket.Put([]byte("l"), block.Hash)

			if err != nil {
				log.Panicln(err) //更新错误
			}

			tip = block.Hash
		} else { //区块链已存在
			tip = bucket.Get([]byte("l"))
		}
		return nil
	})

	if err != nil {
		log.Panicln(err) //更新错误
	}
	bc := Blockchain{tip, db}
	return &bc

}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

func (it *BlockchainIterator) Next() *Block {
	var block Block

	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		encodeBlock := bucket.Get(it.currentHash)
		block = DeSerialize(encodeBlock)
		return nil
	})

	if err != nil {
		log.Panicln(err)
	}
	//指针后移
	it.currentHash = block.PrevBlockHash

	return &block
}

//挖矿添加交易到区块
func (bc *Blockchain) MineBlock(Transactions []*Transaction) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		block := tx.Bucket([]byte(blocksBucket))
		lastHash = block.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panicln(err) //view错误
	}

	newBlock := NewBlock(Transactions, lastHash)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		err := bucket.Put(newBlock.Hash, newBlock.Serialize())

		if err != nil {
			log.Panicln(err) //Put错误
		}

		err = bucket.Put([]byte("l"), newBlock.Hash)

		if err != nil {
			log.Panicln(err) //Put错误
		}
		return nil
	})

	//更新链上的tip
	bc.tip = newBlock.Hash
	if err != nil {
		log.Panicln(err) //更新错误
	}
}

//没有使用的输出的交易集
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	//存储未花费的交易集
	var unspentTXs []Transaction
	//存储已花费的交易ID集
	spentTXOs := make(map[string][]int)

	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			//这笔交易的ID
			txID := hex.EncodeToString(tx.ID)
		Outputs: //这一层遍历交易输出
			for outIdx, out := range tx.Vout {

				//考虑复杂情况,同一交易中,
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						//有多个UTXO,当时有的已经花了，有的没花
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				//能过解锁说明属于未花费的
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			//不是
			if !tx.IsCoinBase() {
				//存在于一笔交易的输入，说明已经花费
				for _, in := range tx.Vin {

					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxID)
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

//没有使用的UTXO
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspendTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspendTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

//没有使用的UTXO,用于构建输入
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated > amount {
					break Work
				}
			}

		}
	}
	return accumulated, unspentOutputs
}
