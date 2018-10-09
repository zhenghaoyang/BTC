package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
}

//创建区块链
func (cli *CLI) createBlockchain(address string) {
	bc := CreatBlockchain(address)
	bc.db.Close()
	fmt.Println("Create Blockchain Done!")
}

func (cli *CLI) getBalance(address string) {

	bc := NewBlockchain(address)
	defer bc.db.Close()
	balance := 0

	UTXOs := bc.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) PrintBlockChain() {
	bc := NewBlockchain("")
	defer bc.db.Close()

	bci := bc.Iterator()
	//var block *Block
	for {
		block := bci.Next()
		fmt.Printf("pre block Hash: %x\n", block.PrevBlockHash)
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

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Printf("send success")

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
			cli.createBlockchain(*creatdBlockchainAddress)
		}
	}
	if printchainCmd.Parsed() {
		cli.PrintBlockChain()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		} else {
			cli.getBalance(*getBalanceAddress)
		}
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Parsed()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
}
func (cli *CLI) createWallet(nodeID string) {

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panicln(err)
	}
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) listAddresses(nodeID string) {
	wallets, _ := NewWallets(nodeID)

	addresses := wallets.GetAllAddress()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
