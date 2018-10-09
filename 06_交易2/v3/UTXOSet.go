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
			fmt.Errorf("CreateBucket err = ", err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			fmt.Errorf("CreateBucket err = ", err)
		}

		return nil
	})
	if err != nil {
		fmt.Errorf(" err = ", err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
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
		fmt.Errorf("FindUTXO db.View err", err)
	}
	return UTXOs
}
