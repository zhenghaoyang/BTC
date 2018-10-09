package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"
)

func main() {
	cli := CLI{}
	cli.Run()
}

//结构体
//Block
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

//序列化
func (b *Block) Serialize() []byte {

	result, err := json.Marshal(&b)
	if err != nil {
		log.Panic(err)
	}
	return result

	//var result bytes.Buffer
	//encoder := gob.NewEncoder(&result)
	//err := encoder.Encode(b)
	//if err != nil {
	//	log.Panic(err)
	//}
	//
	//return result.Bytes()
}

//反序列化 结构体
func DeserializeBlock(d []byte) *Block {
	// //声明一个结构体接收
	var block Block
	err := json.Unmarshal(d, &block)
	if err != nil {
		log.Panic(err)
	}
	return &block
	//var block Block
	//decoder := gob.NewDecoder(bytes.NewReader(d))
	//err := decoder.Decode(&block)
	//if err != nil {
	//	log.Panic(err)
	//}
	//return &block
}

//GenesisBlock
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//新区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//将区块中的交易ID集 hash 产生一个  [32]byte 哈希字节
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	// 遍历 区块的交易集  将区块的每一个交易ID 添加到交易ID的[]byte数组中
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	//对交易集ID hash
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

//ProofOfWork
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

//计算机支持位数
var maxNonce = math.MaxInt64
//构建一个区块完整数据
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

//挖矿 计算hash值
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

//IntToHex int64 ——> []byte
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

//挖矿的难度值
const targetBits = 12

//NewProofOfWork 生成目标难度值 POW
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{
		block,
		target,
	}
	return pow
}

// Validate  PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

//Transaction
type Transaction struct {
	//多个输入输出，就会有多个ID
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// HashPubKey hashes public key
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

//？
// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

//设置交易ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]

}

//TXInput 输入
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
	////标识来自哪个Transaction
	//Txid []byte
	////标识Transaction中哪一个TXInput的Vout
	//// coinbase_Transaction   Vout = -1
	//Vout int
	////解密？
	//ScriptSig string
}

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

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

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

// Lock signs the output
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

//TXOutput UTXO 输出
type TXOutput struct {
	Value      int
	PubKeyHash []byte
	////多少余额
	//Value int
	////属于谁的
	//ScriptPubKey string
}

//解锁脚本 解锁变为交易输入
//func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
//	// 一个scriptSig（解锁脚本），花费者 用自己的私钥 解锁它用于支出
//	return in.ScriptSig == unlockingData
//}

// 锁定脚本 锁定变为交易输出
//func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
//	//用公钥锁定 未来能消费它的
//	return out.ScriptPubKey == unlockingData
//}

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
	return &tx
}

//判断是否属于 coinbase 交易
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// GetWallet returns a Wallet by its address
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
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

//构建新的UTXO交易
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	pubKeyHash := HashPubKey(wallet.PublicKey)

	//找到先前的未花费的UTXOs ->  []TXOutput
	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)
	//余额不足提示
	if acc < amount {
		log.Panic("Error Not enough funds")
	}
	//validOutputs 从构建新的inputs
	for txid, outs := range validOutputs {
		//找的txid
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

	return &tx

}

//一些常量 数据库
const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	//挖矿奖励交易
	subsidy = 10
)

//Blockchain
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

//CreateBlockchain 创建新的Blockchain
func CreateBlockchain(address string) *Blockchain {
	//先判断数据存在
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

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

	bc := Blockchain{tip, db}

	return &bc

}

//NewBlockchain 返回已创建的Blockchain
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

//FindSpendableOutputs 找该地址足够amount的可花费的UTXOs  得先找FindUnspentTransactions 未花费的交易
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	//找该地址未花费交易集
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
Work:
	//遍历全部的花费未交易
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated > amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

//遍历整个区块 找传入地址的所有TX,存入交易集[]Transaction
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	//未花费交易集
	var unspentTXs []Transaction
	//已花费的交易集
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	//从后往前遍历整个区块
	for {
		//区块链上的每一区块
		block := bci.Next()
		//区块交易集中的每一笔交易
		for _, tx := range block.Transactions {
			//每一笔交易ID
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			// tx.Vout 交易中的输出集
			// out --> 输出集中的一笔输出
			//outIdx --> 第几笔输出
			for outIdx, out := range tx.Vout {
				//最新的一个区块上的交易集的每一个交易的多个输出都属于未被花费的UTXO
				//根据交易ID 判断是否属于已花费
				if spentTXOs[txID] != nil {
					for _, spendOut := range spentTXOs[txID] {
						if spendOut == outIdx {
							continue Outputs
						}
					}
				}
				// out --> 输出集中的一笔输出
				// 可以解锁加入未花费的交易集
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			//10 A:Vint[0] ---->TX------>Vout[0]  B  5
			//                   ------>Vout[1]  A  4
			//					------>Vout[2]  C  1

			//属于
			if tx.IsCoinbase() == false {
				//一个交易的输入集
				for _, in := range tx.Vin {
					//可以被解锁，以消费
					if in.UsesKey(pubKeyHash) {
						//找到来自来个上一个区块的交易集的那一个ID,解锁
						//in.Txid 上一个区块的交易集中的一个交易ID
						inTxID := hex.EncodeToString(in.Txid)
						//将上个区块的以消费 inTxID 加入 以消费TXO集合
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs

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

//BlockchainIterator
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//MineBlock
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	//读取最新区块hash
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//新区块
	newBlock := NewBlock(transactions, lastHash)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		//存区块进数据库
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		//存hsah
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		//更新区块链上的 tip hash
		bc.tip = newBlock.Hash

		return nil
	})

}

// Iterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

//Next Iterator
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

//dbExists
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

type CLI struct{}

func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

//validateArgs
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

//printUsage
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	//创建钱包
	wallets := Wallets{}
	//创建钱包字典
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile()

	return &wallets, err
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

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
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

func (cli *CLI) createWallet() {
	//创建钱包
	wallets, _ := NewWallets()
	//通过钱包创建地址
	address := wallets.CreateWallet()
	//钱包存成文件
	wallets.SaveToFile()
	//打印地址
	fmt.Printf("Your new address: %s\n", address)
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

//CLI Run
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
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

//打印区块链
func (cli *CLI) printChain() {
	bc := NewBlockchain("")
	defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
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

//取余额
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

//创建区块链
func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

