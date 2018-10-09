package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func CreatBlockchain(address, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	//先判断数据存在
	if dbExists(dbFile) {
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

func NewBlockchain(nodeID string) *Blockchain {

	dbFile := fmt.Sprintf(dbFile, nodeID)
	fmt.Println("数据库名为", dbFile)
	if dbExists(dbFile) == false {
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
		tip = bucket.Get([]byte("l"))

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

//挖矿添加交易到区块
func (bc *Blockchain) MineBlock(Transactions []*Transaction) {
	var lastHash []byte
	var lastHeight int

	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash = bucket.Get([]byte("l"))
		blockData := bucket.Get(lastHash)
		block := DeSerialize(blockData)
		lastHeight = block.Height
		return nil
	})

	if err != nil {
		log.Panicln(err) //view错误
	}

	newBlock := NewBlock(Transactions, lastHash, lastHeight+1)
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

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	//返回了一个 TransactionID key -> TransactionOutputs 的 map
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			//fmt.Println(tx)
		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}
				//map --  TXOutputs
				outs := UTXO[txID]
				//未花费输出加入 TXOutputs
				outs.Outputs = append(outs.Outputs, out)
				//更新map中的未花费输出
				UTXO[txID] = outs
			}

			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.TxID)
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

func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeSerialize(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Height
}
