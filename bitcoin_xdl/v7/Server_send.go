package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

func sendVersion(addr string, bc *Blockchain) {
	//取得区块高度
	bestHeight := bc.GetBestHeight()
	//装载数据，包含版本，区块高度，节点地址
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	//将请求包装为，version类型的请求
	request := append(commandToBytes("version"), payload...)
	//发送给中心节点
	sendData(addr, request)
}
func sendGetBlocks(address string) {
	//载荷
	payload := gobEncode(getblocks{nodeAddress})
	//包装为getblocks 类型的请求
	request := append(commandToBytes("getblocks"), payload...)
	//发送请求
	sendData(address, request)
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}