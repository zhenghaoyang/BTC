package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

//getblocks 意为 “给我看一下你有什么区块”
type getblocks struct {
	AddrFrom string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}


func sendData(addr string, data []byte) {
	//tcp连接请求
	fmt.Println("test tcp连接请求", addr)
	conn, err := net.Dial(protocol, addr)
	//节点不存在，会报错
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

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}
