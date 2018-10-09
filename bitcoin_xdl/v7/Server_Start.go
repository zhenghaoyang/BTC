package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

// StartServer starts a node
func StartServer(nodeID, minerAddress string) {
	//当前服务节点为
	nodeAddress = fmt.Sprintf("当前服务节点为 localhost:%s", nodeID)
	//指定当前节点为矿工节点
	miningAddress = minerAddress
	//监听中心节点的请求
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	//创建一条区块链
	bc := NewBlockchain(nodeID)
	//不是中心节点，向中心节点发送版本确认
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	//各个节点循环互相监听
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		//处理连接
		go handleConnection(conn, bc)
	}
}

//各个节点的连接处理器
func handleConnection(conn net.Conn, bc *Blockchain) {
	//读取连接的请求
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	//请求头的请求类型，二进制字节转为string命令
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)
	//判断命令的类型
	switch command {
	case "version":
		//版本请求
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

