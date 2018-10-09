package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

func NewBlock(Transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), Transactions, prevBlockHash, nil, 0,height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash[:]
	return block
}

func RootBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{},0)
}

func (block *Block) Serialize() []byte {
	var reuslt bytes.Buffer

	encoder := gob.NewEncoder(&reuslt) //创建编码对象
	err := encoder.Encode(block)       //编码
	if err != nil {
		log.Panicln(err)
	}
	return reuslt.Bytes()
}

func DeSerialize(data []byte) *Block {

	var block Block
	encoder := gob.NewDecoder(bytes.NewReader(data)) //创建解码对象
	err := encoder.Decode(&block)                    //解码
	if err != nil {
		log.Panicln(err)
	}
	return &block
}

func (block *Block) HashTransactions() []byte {
	var transactions [][]byte
	//var txHash [32]byte

	for _, tx := range block.Transactions {
		transactions = append(transactions, tx.ID)

	}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//return txHash[:]
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}
