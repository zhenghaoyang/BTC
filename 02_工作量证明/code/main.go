package main

import (
	"fmt"
	"strconv"
)

func main() {

	bc := NewBlockchain()

	bc.AddBlock("send 1BTC to howy")
	bc.AddBlock("send 2BTC to Ivan")

	for _, block := range bc.blocks {
		fmt.Printf("Prev .hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data  %s\n", block.Data)
		fmt.Printf("Hash  %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
