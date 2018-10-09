package main

import "bytes"


type TXOutput struct {
	Value      int
	PubKeyHash []byte
}


func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}


func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}