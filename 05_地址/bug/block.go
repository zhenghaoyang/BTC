package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}



func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("err=", err)
	}

	return result.Bytes()
}

//反序列化 结构体
func DeserializeBlock(d []byte) *Block {
	//声明一个结构体接收
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}


func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]

}