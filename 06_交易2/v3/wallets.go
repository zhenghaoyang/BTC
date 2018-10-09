package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

//Wallets 用map存储多个Wallet
type Wallets struct {
	Wallets map[string]*Wallet
}

//Wallet 存公私钥对
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

//NewWallet 生成一个Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

//newKeyPair 密钥对
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("newKeyPair err =", err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}



const walletFile = "wallet.dat"
const version = byte(0x00)

// LoadFromFile 从文件中加载wallets
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		fmt.Println("LoadFromFile err ", err)
	}
	var wallets Wallets
	//标识编码格式为椭圆曲线
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)

	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

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

//GetWallet 从多个Wallet取得一个wallet
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

//GetAllAddress 取得全部地址
func (ws *Wallets) GetAllAddress() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
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
