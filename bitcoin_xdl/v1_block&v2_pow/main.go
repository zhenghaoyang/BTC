package main

import (
	"fmt"
	"log"
	"strconv"
)

func main() {
	log.Println("start coding")

	bc := NewBlockchain()

	bc.AddBlock("you get 10 btc")
	bc.AddBlock("you get 10 btc")
	bc.AddBlock("you get 10 btc")
	bc.AddBlock("you get 10 btc")

	for _, block := range bc.blocks {
		fmt.Printf("pre block Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("block data: %s\n", block.Data)
		fmt.Printf("current block Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
