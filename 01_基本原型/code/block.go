package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

//区块结构
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

//创建区块哈希值
func (this *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(this.Timestamp, 10))
	//将一系列[]byte切片连接为一个[]byte切片，之间用sep来分隔，返回生成的新切片
	//将区块数据的三个数据创建成一个切片
	headers := bytes.Join([][]byte{this.PrevBlockHash, this.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	this.Hash = hash[:]
}

////创建区块详细数据
func NewBlock(data string, prevBlockHash []byte) *Block {
	//创建区块
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	//设置哈希
	block.SetHash()
	return block
}

//创始区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
