package main

import (
	"bytes"
	"crypto/sha256"
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
}

//func (b *Block) SetHash() {
//	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
//	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
//	hash := sha256.Sum256(headers)
//	b.Hash = hash[:]
//}

func NewBlock(Transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), Transactions, prevBlockHash, nil, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash[:]
	//block.SetHash()

	return block
}

func RootBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
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
func DeSerialize(data []byte) Block {

	var block Block
	encoder := gob.NewDecoder(bytes.NewReader(data)) //创建解码对象
	err := encoder.Decode(&block)                    //解码
	if err != nil {
		log.Panicln(err)
	}
	return block
}


func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.ID)

	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
