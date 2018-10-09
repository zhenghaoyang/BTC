package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func sendBlock(addr string, b *Block) {
	fmt.Printf("向节点 %v 发送区块\n", addr)
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

//去取得区块
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	//包装为getblocks 类型的请求
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}

//?
func sendData(addr string, data []byte) {
	//fmt.Printf("来自%v的数据请求\n", addr)
	//节点不存在，会报错
	//建立连接
	fmt.Printf("向地址为%v拨号,等待监听连接回传消息\n", addr)
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		//更新节点
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()
	//读取数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address string, kind string, items [][]byte) {
	//?
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)
	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {

	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

//中心节点处理，p2p节点的交易上链
func sendTx(addr string, tnx *Transaction) {
	//打包地址,与交易数据
	data := tx{nodeAddress, tnx.Serialize()}
	//编码
	payload := gobEncode(data)
	//添加tx命令,包装成一个request 让后边的节点识别request类型
	request := append(commandToBytes("tx"), payload...)
	//发送数据请求
	sendData(addr, request)
}

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
