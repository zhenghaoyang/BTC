package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct{}

func (cli *CLI) Run() {

	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("err=", err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("err=", err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("err=", err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("err=", err)
		}

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("err=", err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("err=", err)
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

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

}
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) createWallet() {
	wallets := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Your new address: %s\n", address)
}
func (cli *CLI) listAddresses() {
	wallets := NewWallets()
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		fmt.Println("Address is not valid")
	}

	//bc := NewBlockchain(address)
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("createBlockchain Done")

}

func (cli *CLI) printChain() {
	bc := NewBlockchain("")
	defer bc.db.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Block %x \n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

}

//发送不成功
func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
	}
	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("send Success!")

}

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}


//FindUTXO
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	//找到所有的未花费交易集
	unspendTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspendTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}