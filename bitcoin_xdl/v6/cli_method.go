package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

//创建区块链
func (cli *CLI) createBlockchain(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := CreatBlockchain(address, nodeID)
	bc.db.Close()
	fmt.Println("Create Blockchain Done!")
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

func (cli *CLI) getBalance(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()
	defer bc.db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) listAddresses(nodeID string) {
	wallets, _ := NewWallets(nodeID)

	addresses := wallets.GetAllAddress()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) PrintBlockChain(nodeID string) {
	bc := NewBlockchain(nodeID)
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

func (cli *CLI) send(from, to string, nodeID string, amount int) {

	if !ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(nodeID)
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	UTXOSet := UTXOSet{bc}
	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	bc.MineBlock([]*Transaction{tx})
	fmt.Printf("send success")

}
