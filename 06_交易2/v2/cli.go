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

	//二级提示
	createBlockchainAddress := createBlockchainCmd.String("addr", "", "The address to send genesis block reward to")

	switch os.Args[1] {
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Errorf("createWalletCmd.Parse", err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Errorf("listAddressesCmd.Parse", err)
		}

	case "createBC":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Errorf("createBlockchainCmd.Parse", err)
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

}
