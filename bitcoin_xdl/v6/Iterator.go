package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
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
