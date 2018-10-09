package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/ripemd160"
	"log"
	"math/big"
	"os"
	"time"
)

const walletFile = "wallet.dat"
const addressChecksumLen = 4
const version = byte(0x00)
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

const targetBits = 12
const subsidy = 10

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	ReverseBytes(result)

	for b := range input {
		if b == 0x00 {
			result = append([]byte{b58Alphabet[0]}, result...)
		} else {
			break
		}
	}

	return result
}

// Base58Decode decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 0

	for b := range input {
		if b == 0x00 {
			zeroBytes++
		}
	}

	payload := input[zeroBytes:]
	for _, b := range payload {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)

	return decoded
}

//NewCoinbaseTX 挖矿产生的交易
func NewCoinbaseTX(to, data string) *Transaction {

	//矿工可以给这笔交易添加数据
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	//交易输入
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	//交易输出
	txout := NewTXOutput(subsidy, to)
	//创建交易信息
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.SetID()
	fmt.Println("-------------------NewCoinbaseTX-----------------------------")
	fmt.Println(tx)
	return &tx
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc

}

func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Println("err=", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		fmt.Println("err=", err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{block, target}
	return pow

}

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		fmt.Println("err=", err)
	}
	return buff.Bytes()

}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0

}

//创建钱包
func NewWallets() *Wallets {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	// 第二次应从文件取
	err := wallets.LoadFromFile()
	if err != nil {
		fmt.Println("wallets load from fail err=", err)
	}

	return &wallets
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func checksum(versionPayload []byte) []byte {
	firstSHA := sha256.Sum256(versionPayload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		fmt.Println("err=", err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

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
		fmt.Println("GenerateKey private err = ", err)
	}
	PubKey := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, PubKey
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets := NewWallets()
	wallet := wallets.GetWallet(from)

	pubKeyHash := HashPubKey(wallet.PublicKey)
	//找到先前的未花费的UTXOs ->  []TXOutput
	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

	//余额不足提示
	if acc < amount {
		fmt.Println("Error Not enough funds")
	}

	//validOutputs 从构建新的inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}
	// Build a list of outputs

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change

	}
	tx := Transaction{nil, inputs, outputs}

	tx.SetID()
	bc.SignTransaction(&tx, wallet.PrivateKey)
	//	fmt.Println(tx) ok
	return &tx
}
