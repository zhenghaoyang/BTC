package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

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
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		fmt.Println(txID)
		for outIdx, out := range tx.Vout {
			//验证发送者给的锁，即该笔交易属于我,花
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

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
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
	return unspentTXs
}
func (bc *Blockchain) MineBlock(transactions []*Transaction) {

	var lastHash []byte

	for _, tx := range transactions {
		//bc.VerifyTransaction(tx) =true
		fmt.Println(bc.VerifyTransaction(tx))
		if bc.VerifyTransaction(tx) != true {
			//log.Panic(err)
			fmt.Errorf("bc.VerifyTransaction(tx)")
			//return
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		fmt.Println("bc.db.View err", err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		err := b.Put(newBlock.Hash, newBlock.Serialize())

		if err != nil {
			fmt.Println("bc.db.Update err", err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			//log.Panic(err)
			fmt.Println("err ")
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		fmt.Println("MineBlock err ", err)
	}

}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {

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
	//return ture
	return tx.Verify(prevTXs)

}
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
