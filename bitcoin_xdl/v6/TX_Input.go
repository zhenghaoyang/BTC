package main

import "bytes"

type TXInput struct {
	TxID      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}


// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}


////解锁脚本
//func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
//	// 一个scriptSig（解锁脚本），花费者 用自己的私钥 解锁它用于支出
//	return in.ScriptSig == unlockingData
//}
