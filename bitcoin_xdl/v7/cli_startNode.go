package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(nodeID, minerAddress string) {

	fmt.Printf("Starting node %s\n", nodeID)
	//挖矿地址 有传值 minerAddress
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	//开启节点服务, minerAddress 有传值 为挖矿节点
	StartServer(nodeID, minerAddress)
}

