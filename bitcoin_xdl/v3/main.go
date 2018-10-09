package main

import (
	"log"
)

func main() {
	log.Println("mining btc starting")

	bc := NewBlockchain()
	defer bc.db.Close()

	cli := CLI{bc}

	cli.Run()
	//
	//bc.AddBlock("you get 10 btc")
	//bc.AddBlock("you get 10 btc")
	//bc.AddBlock("you get 10 btc")
	//bc.AddBlock("you get 10 btc")
	//
	//for _, block := range bc.blocks {
	//	fmt.Printf("pre block Hash: %x\n", block.PrevBlockHash)
	//	fmt.Printf("block data: %s\n", block.Data)
	//	fmt.Printf("current block Hash: %x\n", block.Hash)
	//	pow := NewProofOfWork(block)
	//	fmt.Printf("pow %s\n", strconv.FormatBool(pow.Validate()))
	//	fmt.Println()
	//}
}
