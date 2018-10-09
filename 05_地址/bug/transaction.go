package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		fmt.Println("err=", err)
	}

	return encoded.Bytes()
}

//Error
// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			fmt.Println("ERROR: Previous transaction is not correct")
			return false
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {

		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]

		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(vin.Signature)

		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}

		keyLen := len(vin.PubKey)

		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		//校验签名
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			fmt.Println("test Verify return false", ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s))
			return false
		}

	}
	fmt.Println("test Verify return true")
	return true

}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

func (tx *Transaction) SetID() {
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

func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction %x", tx.ID))
	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
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

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	//遍历交易输出，构建交易
	for inID, vin := range txCopy.Vin {
		//找到之前的一笔交易
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		//之前的一笔交易的输出公钥Hash付给当前输入
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		txCopy.ID = txCopy.Hash()
		//手动回收
		txCopy.Vin[inID].PubKey = nil
		//私钥产生用于签名的r,s
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		//产生签名
		signature := append(r.Bytes(), s.Bytes()...)
		//签名
		tx.Vin[inID].Signature = signature
	}
}
