package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"
const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

//Wallets 用map存储多个Wallet
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile(nodeID)

	return &wallets, err
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile(nodeID string) error {
	//重命名钱包文件
	walletFile := fmt.Sprintf(walletFile, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		fmt.Printf("遇见err,But I will let you go")
		return nil
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

//NewWallet 生成一个Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}
//GetWallet 从多个Wallet取得一个wallet
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// GetAddress 取得wallet一个地址
func (w Wallet) GetAddress() []byte {
	//公钥哈希
	pubKeyHash := HashPubKey(w.PublicKey)
	//过渡版本
	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	//新的公私钥
	wallet := NewWallet()
	//公私钥生成---->地址
	address := fmt.Sprintf("%s", wallet.GetAddress())
	//k ---- 地址, v----wallet
	ws.Wallets[address] = wallet
	return address
}

//GetAllAddress 取得全部地址
func (ws *Wallets) GetAllAddress() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer

	walletFile := fmt.Sprintf(walletFile, nodeID)

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
