package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)



type Wallets struct {
	Wallets map[string]*Wallet
}

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (this Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(this.PublicKey)
	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)
	fullPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullPayload)
	return address
}

//创建一个钱包地址 Wallet
func (ws Wallets) CreateWallet() string {
	fmt.Println("CreateWallet")
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}

func (this *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		fmt.Println("err=", err)
	}
	var wallets Wallets
	//标识
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))

	err = decoder.Decode(&wallets)
	if err != nil {
		fmt.Println("err=", err)
	}
	this.Wallets = wallets.Wallets
	return nil
}

func (this *Wallets) SaveToFile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(this)
	if err != nil {
		fmt.Println("err=", err)
	}
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		fmt.Println("err=", err)
	}

}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}
