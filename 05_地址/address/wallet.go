package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	//生成公私对
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

//GetAddress return wallet address
func (w Wallet) GetAddress() []byte {
	//双哈希
	pubKeyHash := HashPubKey(w.PublicKey)

	//先把 version号加进去
	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)
	fullPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullPayload)
	return address

}

func ValidateAddress(address string)bool{
	//公钥hash 解码
	pubKeyHash := Base58Decode([]byte(address))
	//末4尾 校验码
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	//首位 版本号
	version := pubKeyHash[0]
	//中间 公钥hash
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-addressChecksumLen]
	//checksum 双hash，变为地址
	targetCheckSum := checksum(append([]byte{version},pubKeyHash...))
	//只对比末四位？
	return bytes.Compare(actualChecksum,targetCheckSum) == 0
}


func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}

//对私钥生成的公钥 双hash
func HashPubKey(pubKey []byte) []byte {

	publicSHA256 := sha256.Sum256(pubKey)
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])

	if err != nil {
		log.Panic(err)
	}

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

