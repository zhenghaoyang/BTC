package main

import (
	"fmt"
)

//listAddresses 列出全部地址
func (cli *CLI) listAddresses() {
	wallets := NewWallets()

	addresses := wallets.GetAllAddress()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
