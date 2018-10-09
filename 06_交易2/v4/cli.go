package main

import (
	"flag"
	"fmt"
	"os"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  createBC -addr ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")

}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createBC", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)


	//二级提示
	createBlockchainAddress := createBlockchainCmd.String("addr", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")


	switch os.Args[1] {
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("createWalletCmd.Parse", err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("listAddressesCmd.Parse", err)
		}

	case "createBC":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("createBlockchainCmd.Parse", err)
		}

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("createBlockchainCmd.Parse", err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("createBlockchainCmd.Parse", err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("sendCmd.Parse", err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

}
