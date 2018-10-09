package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 16

type ProofOfWork struct {
	block  *Block
	target *big.Int //去指针
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{block, target}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce)),
	}, []byte{})
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("current block data: %s", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash) //\r回到开头位置
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce ++
		}
	}
	fmt.Println()
	fmt.Println()
	return nonce, hash[:]

}
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	var hash [32]byte
	data := pow.prepareData(pow.block.Nonce)
	hash = sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}
