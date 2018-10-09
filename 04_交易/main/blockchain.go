package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"


type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

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

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//FindSpendableOutputs
func (bc *Blockchain)FindSpendableOutputs(address string,amount int) (int ,map[string][]int){
		unspentOutputs := make(map[string][]int)
		unspentTXs := bc.FindUnspentTransactions(address)
		accumulated := 0
		Work:
			for _,tx := range unspentTXs{
				txID := hex.EncodeToString(tx.ID)
				for outIdx,out := range tx.Vout{
					if out.CanBeUnlockedWith(address) && accumulated <amount {
						accumulated += out.Value
						unspentOutputs[txID] = append(unspentOutputs[txID],outIdx)
						if accumulated > amount{
							break Work
						}
					}
				}
			}


		return accumulated, unspentOutputs
}


//10 A:Vint[0] ---->TX------>Vout[0] B 5
                   //------>Vout[1]  A  4

//遍历区块 找传入地址的所有UTXO,存入交易数组[]Transaction
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	//未花费交易集
	var unspentTXs []Transaction
	//已花费的
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for{
		block := bci.Next()

		for _,tx := range block.Transactions {
			//这笔交易的ID
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			// tx.Vout 是
			// out ==>
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spendOut := range spentTXOs[txID] {
						//spendOut 索引
						if spendOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
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

	return unspentTXs

}


func CreateBlockchain(address string)  *Blockchain{
	//先判断数据存在
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db,err := bolt.Open(dbFile,0600,nil)
	if err!=nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address,genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b,err:= tx.CreateBucket([]byte(blocksBucket))
		if err!= nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash,genesis.Serialize())
		if err!= nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"),genesis.Hash)
		if err!=nil {
			log.Panic(err)
		}

		tip = genesis.Hash
		return nil

	})


	bc := Blockchain{tip, db}

	return &bc

}


func NewBlockChain(address string) *Blockchain{
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


//FindUTXO
func (bc *Blockchain)FindUTXO(address string)[]TXOutput{
	var UTXOs []TXOutput
	unspendTransactions := bc.FindUnspentTransactions(address)

	for _,tx:= range unspendTransactions{
		for _,out := range tx.Vout{
			if out.CanBeUnlockedWith(address){
				UTXOs = append(UTXOs,out)
			}
		}
	}
	return UTXOs
}
