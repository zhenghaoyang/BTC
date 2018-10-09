package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	cli := CLI{}
	cli.Run()
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	//根据nodeID创建BC
	bc := NewBlockchain(nodeID)
	//创建UTXOset
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	//指定当前节点挖矿
	if mineNow {
		fmt.Printf("当前节点挖矿为 %v\n", nodeID)
		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		//没有指定发送给中心节点处理
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}

func (cli *CLI) startNode(nodeID, minerAddress string) {

	fmt.Printf("Starting node %s\n", nodeID)
	//挖矿地址 有传值 minerAddress
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("开始挖矿----------------------------")
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	//开启节点服务, minerAddress 有传值 为挖矿节点
	StartServer(nodeID, minerAddress)
}

// StartServer starts a node
func StartServer(nodeID, minerAddress string) {
	//当前服务节点为当前服务节点为
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
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
		fmt.Printf("nodeID %v 正在监听\n", nodeID)
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		//处理连接
		go handleConnection(conn, bc)
	}
}
