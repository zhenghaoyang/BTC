package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 50

type TXInput struct {
	TxID      []byte
	Vout      int
	ScriptSig string
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) IsCoinBase() bool {
	//一个输入，
	// 输入的上个交易id为nil,来自系统，故Vin[0]交易ID长度=0
	//没有索引，这个交易的输入来自上个交易的那个输出，并不存在 故为-1

	return len(tx.Vin) == 1 && len(tx.Vin[0].TxID) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panicln(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// 锁定脚本
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	//用公钥锁定 未来能消费它的
	return out.ScriptPubKey == unlockingData
}

//解锁脚本
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	// 一个scriptSig（解锁脚本），花费者 用自己的私钥 解锁它用于支出
	return in.ScriptSig == unlockingData
}

//挖矿交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprint("Reward to %s", to)
	}

	txVin := TXInput{[]byte{}, -1, data}
	txVout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txVin}, []TXOutput{txVout}}
	tx.SetID()
	return &tx

}

//转账交易
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	//找出UTXO集
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panicln("not enought UTXO")
	}
	//交易输入
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panicln(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}
	//交易输出
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
