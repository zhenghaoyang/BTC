package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

func handleVersion(request []byte, bc *Blockchain) {
	//开辟内存
	var buff bytes.Buffer
	//数据
	var payload verzion
	///将请求头,写进buff,不包含请求类型
	buff.Write(request[commandLength:])
	//编码
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//取得当前节点区块链的区块高度
	myBestHeight := bc.GetBestHeight()
	//另外节点发来版本数据的区块高度
	foreignerBestHeight := payload.BestHeight
	//当本节点区块高度小于中心节点的高度，
	if myBestHeight < foreignerBestHeight {
		//发送请求区块数据的请求
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		//大于发送请求版本
		fmt.Println("test 本区块高度大于其他区块发送请求版本确认")
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}
