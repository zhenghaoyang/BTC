package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

const utxoBucket = "chainstate"

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)

		if err != nil && err != bolt.ErrBucketNotFound {
			fmt.Println("Reindex db.DeleteBucket err ", err)
		}
		_, err = tx.CreateBucket(bucketName)

		if err != nil {
			fmt.Println("Reindex tx.CreateBucket err ", err)
		}

		return nil
	})
	if err != nil {
		fmt.Println("Reindex db.Update err ", err)
	}

	//重建UTXO
	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				fmt.Println("Reindex DecodeString err ", err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				fmt.Println("Reindex db.put err ", err)
			}
		}

		return nil
	})
	if err != nil {
		fmt.Println("Reindex db.Update err ", err)
	}

}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("FindUTXO db.View err", err)
	}
	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("FindSpendableOutputs db.View err", err)
	}
	return accumulated, unspentOutputs
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {

	db := u.Blockchain.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				//通过当前交易的输入集 去取 UTXO
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					for outIndex, out := range outs.Outputs {
						if outIndex != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					//UTXO 数量为零,说明此交易可以从数据库 删除
					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							fmt.Println("b.Delete err", err)
						}
					} else { //还有UTXO, 更新数据库的UTXO
						err := b.Put(vin.Txid, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}
			//新区块的新的UTXO
			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				fmt.Println("b.Put err", err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(" db.Update err", err)

	}
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0


	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter

}