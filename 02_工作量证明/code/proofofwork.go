package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//挖矿的难度值
const targetBits = 8

//包含要处理的两个数据
type ProofOfWork struct {
	block *Block
	//目标哈希
	target *big.Int
}

//生成目标难度值
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		block,
		target,
	}
	return pow
}

//准备数据 来进行哈希
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	//Join将一系列[]byte切片连接为一个[]byte切片，之间用sep来分隔，返回生成的新切片
	data := bytes.Join(
		// byte 数组
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

var (
	maxNonce = math.MaxInt64
)

//pow工作,计算符合目标值的哈希值
func (pow *ProofOfWork) Run() (int, []byte) {
	// 用于接受 数据处理完后的哈希值
	var hashInt big.Int
	var hash [32]byte
	//用于生成变化哈希值的参数
	nonce := 0
	fmt.Printf("Mining the block containing %s\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		// hash := sha256.Sum256(data) 导致hash 0000....
		hash = sha256.Sum256(data)
		fmt.Printf("第%d次碰撞  \n", nonce)
		// hashInt
		// sets z to that value, and returns z.
		hashInt.SetBytes(hash[:])

		//hashInt > pow.target
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:]
}

//在处理
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
