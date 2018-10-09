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
