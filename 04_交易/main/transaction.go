package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	//多个输入输出，就会有多个ID
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Txid      []byte
	//Vout 索引  标识来自哪个Output  coinbase  Vout = -1
	Vout      int
	ScriptSig string
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}


//解锁脚本
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool{
	// 一个scriptSig（解锁脚本），花费者 用自己的私钥 解锁它用于支出
	return in.ScriptSig == unlockingData
}


// 锁定脚本
func (out *TXOutput) CanBeUnlockedWith(unlockingData string)  bool {
	//用公钥锁定 未来能消费它的
	return out.ScriptPubKey == unlockingData
}


func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//设置交易ID
func (tx *Transaction)SetID(){
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)

	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]

}


//挖矿奖励交易
const subsidy = 10
func NewCoinbaseTX(to, data string) *Transaction {

	//矿工可以给这笔交易添加数据
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	//交易输入
	txin := TXInput{[]byte{}, -1, data}
	//交易输出
	txout := TXOutput{subsidy, to}
	//创建交易信息
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	return &tx
}


func NewUTXOTransaction(from,to string,amount int,bc *Blockchain) *Transaction{
	var inputs []TXInput
	var outputs []TXOutput

	acc,validOutputs := bc.FindSpendableOutputs(from,amount)
	if acc < amount{
		log.Panic("Error Not enough funds")
	}
	for txid,outs := range validOutputs {
		txID,err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
			for _,out := range outs{
				input := TXInput{txID,out,from}
				inputs = append(inputs,input)
		}
	}
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs,TXOutput{acc-amount,from})
	}
	tx := Transaction{nil,inputs,outputs}
	tx.SetID()
	return &tx

}


