package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

const subsidy = 50

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

//挖矿交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprint("Reward to %s", to)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()
	return &tx

}

//打印交易完整信息
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TxID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

//转账交易
func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := HashPubKey(wallet.PublicKey)

	//找出UTXO集
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

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
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}
	//交易输出
	// Build a list of outputs
	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	//todo
	//UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}
