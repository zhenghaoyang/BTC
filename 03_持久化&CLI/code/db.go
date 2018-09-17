package main

import (
	_ "github.com/boltdb/bolt"
)

func init() {
	const dbFile = "blockchain.db"
	const blocksBucket = "blocks"
}
