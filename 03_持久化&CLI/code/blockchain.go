package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

//整个数据库
const dbFile = "blockchain.db"

//多个区块
const blocksBucket = "blocks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//添加区块
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	//依据上个区块，即当前最后一个区块的哈希，创建新的区块
	newBlock := NewBlock(data, lastHash)

	//添加新区块
	err = bc.db.Update(func(tx *bolt.Tx) error {
		//
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())

		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		//更新区块链的最后一个区块的哈希
		bc.tip = newBlock.Hash
		return nil
	})
}

// 遍历整个区块链的迭代器
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

//迭代器的方法
func (i *BlockchainIterator) Next() *Block {

	var block *Block

	// tx *bolt.Tx 事务 View 只读权限
	err := i.db.View(func(tx *bolt.Tx) error {
		//用于存放blocks的桶，重新创建的？
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	//遍历blockChain时，需要不断的将当前hash传给下个block
	i.currentHash = block.PrevBlockHash
	return block
}

//创建区块链
func NewBlockchain() *Blockchain {

	var tip []byte
	//数据库
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			//用于存放blocks的桶
			b, err := tx.CreateBucket([]byte(blocksBucket))

			if err != nil {
				log.Panic(err)
			}

			err = b.Put(genesis.Hash, genesis.Serialize())

			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"), genesis.Hash)

			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			//最后跟新末尾区块的hash
			tip = b.Get([]byte("l"))
		}
		return nil

	})

	if err != nil {
		log.Panic(err)
	}

	//依据末尾hash &数据库创建 区块链
	bc := Blockchain{tip, db}
	return &bc

}
