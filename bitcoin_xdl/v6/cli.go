package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
}

func (cli *CLI) Run() {
	cli.validateArgs()

	//Getenv检索并返回名为key的环境变量的值。如果不存在该环境变量会返回空字符串
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		//不退出
		os.Exit(1)
	}

	createblockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	creatdBlockchainAddress := createblockchainCmd.String("address", "", "输入接受奖励的地址")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount")

	switch os.Args[1] {
	case "createblockchain":
		err := createblockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)
		}

	case "printchain":
		err := printchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)

		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln(err)

		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Errorf("createWalletCmd.Parse", err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if createblockchainCmd.Parsed() {
		if *creatdBlockchainAddress == "" {
			createblockchainCmd.Usage()
			os.Exit(1)
		} else {
			cli.createBlockchain(*creatdBlockchainAddress,nodeID)
		}
	}
	if printchainCmd.Parsed() {
		cli.PrintBlockChain(nodeID)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		} else {
			cli.getBalance(*getBalanceAddress,nodeID)
		}
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Parsed()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, nodeID, *sendAmount)
		//cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
}
