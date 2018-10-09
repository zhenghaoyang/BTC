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
	"strings"
	"time"
)

func main() {
	cli := CLI{}
	cli.Run()
}

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
		fmt.Println(*sendFrom)
		fmt.Println(*sendTo)
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
	fmt.Println("outputs")
	fmt.Println(outputs)
	fmt.Println("outputs[0]")
	fmt.Println(outputs[0].Value)
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
		fmt.Println("outputs[1]")
		fmt.Println(outputs[1].Value)
	}
	tx := Transaction{nil, inputs, outputs}

	tx.SetID()
	bc.SignTransaction(&tx, wallet.PrivateKey)
	fmt.Println("-------------------NewUTXOTransaction-----------------------------")
	fmt.Println("-------------------NewUTXOTransaction-----------------------------")
	fmt.Println("-------------------NewUTXOTransaction-----------------------------")
	fmt.Println("-------------------NewUTXOTransaction-----------------------------")
	fmt.Println(tx)
	return &tx

}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	//遍历当前交易的输入集
	for _, vin := range tx.Vin {
		//找到之前的交易ID
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		//之前的交易集
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	//遍历交易输出，构建交易
	for inID, vin := range txCopy.Vin {
		//找到之前的一笔交易
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		//之前的一笔交易的输出公钥Hash付给当前输入
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		txCopy.ID = txCopy.Hash()
		//手动回收
		txCopy.Vin[inID].PubKey = nil
		//私钥产生用于签名的r,s
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		//产生签名
		signature := append(r.Bytes(), s.Bytes()...)
		//签名
		tx.Vin[inID].Signature = signature
	}
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

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil {
		fmt.Println("Next err=", err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

// Iterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

//FindSpendableOutputs 找该地址足够amount的可花费的UTXOs  得先找FindUnspentTransactions 未花费的交易
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
	ttt := ""
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		fmt.Println(txID)
		for outIdx, out := range tx.Vout {
			//验证发送者给的锁，即该笔交易属于我,花
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {

				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				ttt = txID
				if accumulated > amount {
					break Work
				}
			}
		}
	}
	fmt.Println("	unspentOutputs[txID]")
	fmt.Println("	unspentOutputs[0]")
	fmt.Println("	unspentOutputs[0]")
	//
	fmt.Println(ttt)
	fmt.Println(unspentOutputs[ttt])

	return accumulated, unspentOutputs

}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			fmt.Println("FindUnspentTransactions")
			fmt.Println("FindUnspentTransactions")
			fmt.Println("FindUnspentTransactions")
			fmt.Println(tx)
		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
	fmt.Println("	unspentTXs[0]")
	fmt.Println("	unspentTXs[0]")
	fmt.Println("	unspentTXs[0]")

	fmt.Println(unspentTXs[0])
	return unspentTXs
}
func (bc *Blockchain) MineBlock(transactions []*Transaction) {

	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			//log.Panic(err)
			fmt.Println("ERROR: Invalid transaction")
			//return
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		s := hex.EncodeToString(lastHash)
		fmt.Println("b.Get([]byte", s)
		return nil
	})

	if err != nil {
		fmt.Println("bc.db.View err", err)
	}

	newBlock := NewBlock(transactions, lastHash)
	fmt.Println("newBlock := NewBlock(transactions, lastHash)")
	fmt.Println(newBlock.Transactions[0])
	//fmt.Println(newBlock.Transactions[1])

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())

		if err != nil {
			fmt.Println("bc.db.Update err", err)
		}
		//哈哈哈
		err = b.Put([]byte("l"), newBlock.Hash)


		bc.tip = newBlock.Hash
		str := hex.EncodeToString(newBlock.Hash)
		fmt.Println("newBlock.Hash", str)
		return nil
	})
	if err != nil {
		fmt.Println("MineBlock err ", err)
	}

}
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	//fmt.Println("----------------------VerifyTransaction----------------------")
	//fmt.Println("----------------------VerifyTransaction----------------------")
	//fmt.Println("----------------------VerifyTransaction----------------------")
	//fmt.Println(tx)
	//if tx.IsCoinbase() {
	//	return true
	//}

	prevTXs := make(map[string]Transaction)

	//遍历当前交易的输入
	for _, vin := range tx.Vin {
		//通过当前输入的交易id 找到 来自那个交易
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			fmt.Println("FindTransaction err = ", err)
			log.Panic(err)
		}
		//之前的交易集
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)

}
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {

	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				fmt.Println("found FindTransaction ")
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
}
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
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

const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

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

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("err=", err)
	}

	return result.Bytes()
}

//反序列化 结构体
func DeserializeBlock(d []byte) *Block {
	//声明一个结构体接收
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]

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

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func (pow *ProofOfWork) Run() (int, []byte) {
	maxNonce := math.MaxInt64
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining------a new block---------")
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
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransaction(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		fmt.Println("err=", err)
	}
	return buff.Bytes()

}

const targetBits = 12

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{block, target}
	return pow

}

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		fmt.Println("err=", err)
	}

	return encoded.Bytes()
}

//Error
// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {

	fmt.Println(" tx.IsCoinbase() ", tx.IsCoinbase())
	//false
	if tx.IsCoinbase() {
		fmt.Println("false")
		fmt.Println("false")
		fmt.Println("false")
		return true
	}

	fmt.Println(" tx.Vin  ")
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			fmt.Println(" tx.Vin  ", false)
			fmt.Println("ERROR: Previous transaction is not correct")
			return false
		}
	}

	fmt.Println(" tx.Vin  ok ")

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {

		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]

		fmt.Println("Verify")
		fmt.Println("Verify")
		fmt.Println("Verify")

		fmt.Println(prevTX)
		fmt.Println("=======================================")
		fmt.Println("=======================================")
		fmt.Println("=======================================")
		fmt.Println(tx)

		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(vin.Signature)

		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}

		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		//fmt.Println("ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false", ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s))
		//校验签名
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}

	}
	return true

}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}
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
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction %x", tx.ID))
	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
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

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

const version = byte(0x00)

func (this Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(this.PublicKey)
	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)
	fullPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullPayload)
	return address
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

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
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

const addressChecksumLen = 4

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

type Wallets struct {
	Wallets map[string]*Wallet
}

//创建一个钱包地址 Wallet
func (ws Wallets) CreateWallet() string {
	fmt.Println("CreateWallet")
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
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

const walletFile = "wallet.dat"

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
