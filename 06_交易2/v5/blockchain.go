package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second "

// Blockchain 区块链
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {

	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	//创世交易
	cbtx := FirsCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	//打开数据库
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		fmt.Println("bolt.Open err = ", err)
	}

	//插入数据
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			fmt.Println("bolt.Open err = ", err)
		}
		//插入区块数据
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			fmt.Println("bolt.Open err = ", err)
		}
		//区块链末尾hash
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			fmt.Println("bolt.Open err = ", err)
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

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
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

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	//返回了一个 TransactionID -> TransactionOutputs 的 map
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

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
				//map --  TXOutputs
				outs := UTXO[txID]
				//未花费输出加入 TXOutputs
				outs.Outputs = append(outs.Outputs, out)
				//更新map中的未花费输出
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
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

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTranscation(vin.Txid)
		if err != nil {
			fmt.Println("FindTranscation err", err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTranscation(ID []byte) (Transaction, error) {
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

func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			fmt.Println("VerifyTransaction err")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		fmt.Println("MineBlock bc.db.View err", err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			fmt.Println("MineBlock bc.db.Update err", err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			fmt.Println("MineBlock bc.db.Put err", err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		fmt.Println("MineBlock err", err)
	}

	return newBlock

}

// VerifyTransaction verifies transaction input signatures 验证每笔交易输入的签名
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTranscation(vin.Txid)
		if err != nil {
			fmt.Println("VerifyTransaction err")
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
