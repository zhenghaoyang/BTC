package main

import (
	"encoding/json"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

//序列化
func (b *Block) Serialize() []byte {

	result, err := json.Marshal(&b)
	if err != nil {
		log.Panic(err)
	}
	return result

	// //定义一个缓冲区
	// var result bytes.Buffer
	// //包装 缓冲区
	// encoder := gob.NewEncoder(&result)
	// //将block 编码 --> 序列号
	// err := encoder.Encode(b)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// return result.Bytes()
}

//反序列化 结构体
func DeserializeBlock(d []byte) *Block {
	// //声明一个结构体接收
	var block Block
	err := json.Unmarshal(d, &block)
	if err != nil {
		log.Panic(err)
	}
	return &block
	// //声明一个结构体接收
	// var block Block
	// decoder := gob.NewDecoder(bytes.NewReader(d))
	// err := decoder.Decode(&block)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// return &block
}

// NewBlock creates and returns Block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
