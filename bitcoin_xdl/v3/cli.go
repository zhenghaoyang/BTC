package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	blockchain *Blockchain
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
func (cli *CLI) printUsage() {
	fmt.Println("addblock add block to blockchain")
	fmt.Println("printchain PrintBlockChain")
}

func (cli *CLI) AddBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("add block success")
}
func (cli *CLI) PrintBlockChain() {
	bci := cli.blockchain.Iterator()
	//var block *Block
	for {
		block := bci.Next()
		fmt.Printf("pre block Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("block data: %s\n", block.Data)
		fmt.Printf("current block Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		//创世块结束循环
 		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()
	addblockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addblockCmd.String("data", "", "输入添加区块的数据")
	switch os.Args[1] {
	case "addblock":
		err := addblockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}

	case "printchain":
		err := printchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)

		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if addblockCmd.Parsed() {
		if *addBlockData == "" {
			addblockCmd.Usage()
			os.Exit(1)
		} else {
			cli.AddBlock(*addBlockData)
		}
	}
	if printchainCmd.Parsed() {
		cli.PrintBlockChain()
	}
}
