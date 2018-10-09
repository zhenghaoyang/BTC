package main

import "fmt"

func (cli *CLI) createWallet() {

	wallets := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}



// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() *Wallets {

	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile()

	if err!=nil {
		fmt.Errorf("LoadFromFile err",err)
	}

	return &wallets
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}