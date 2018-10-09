package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
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

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		block := tx.Bucket([]byte(blocksBucket))
		lastHash = block.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panicln(err) //view错误
	}

	newBlock := NewBlock(data, lastHash)
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

func NewBlockchain() *Blockchain {
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
			block := RootBlock(genesisCoinbaseData)
			err = bucket.Put(block.Hash, block.Serialize())

			if err != nil {
				log.Panicln(err) //更新错误
			}

			err = bucket.Put([]byte("l"), block.Hash)

			if err != nil {
				log.Panicln(err) //更新错误
			}

			tip = block.Hash
		} else {
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
