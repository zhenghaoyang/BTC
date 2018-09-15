package main

import (
	"fmt"
)

func main() {

	bc := NewBlockchain()

	bc.AddBlock("send 1BTC to howy")
	bc.AddBlock("send 2BTC to Ivan")

	for _, block := range bc.blocks {
		fmt.Printf("Prev .hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data  %s\n", block.Data)
		fmt.Printf("Hash  %X\n", block.Hash)
	}
}
