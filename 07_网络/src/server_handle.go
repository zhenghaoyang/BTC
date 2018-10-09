package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

func handleConnection(conn net.Conn, bc *Blockchain) {
	//
	fmt.Printf("读取Request 提取Command\n")
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	//请求头的请求类型，二进制字节转为string命令
	command := bytesToCommand(request[:commandLength])
	//打印出当前节点接收到命令
	fmt.Printf("Received %s command\n", command)
	//判断命令的类型
	switch command {
	case "addr":
		//处理节点
		handleAddr(request)
	case "block":
		//处理block请求
		handleBlock(request, bc)
	case "inv":
		//处理清单请求
		handleInv(request, bc)
	case "getblocks":
		//
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		//版本请求
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

//当接收到一个新块时，我们把它放到区块链里面。
//如果还有更多的区块需要下载，我们继续从上一个下载的块的那个节点继续请求。
//当最后把所有块都下载完后，对 UTXO 集进行重新索引。
func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//取得payload中的Block
	blockData := payload.Block
	block := DeserializeBlock(blockData)
	//节点block添加到区块链上
	bc.AddBlock(block)
	fmt.Printf("Added block %x\n", block.Hash)
	//区块中交易大于一个
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		//发送区块中的数据
		sendGetData(payload.AddrFrom, "block", blockHash)
		//跟新区块中的交易集
		blocksInTransit = blocksInTransit[1:]
	} else {
		//重建索引
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}

}

//收到块哈希，我们想要将它们保存在 blocksInTransit 变量来跟踪已下载的块。
//这能够让我们从不同的节点下载块。在将块置于传送状态时，
//我们给 inv 消息的发送者发送 getdata 命令并更新 blocksInTransit
func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)
	//处理inv的不同类型
	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0] //取出第一个block的hash
		sendGetData(payload.AddrFrom, "block", blockHash)
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			//不再内存池中
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}
	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}

}

//返回 所有块哈希
func handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	//去得全部区块
	blocks := bc.GetBlockHashes() //发送清单
	sendInv(payload.AddrFrom, "block", blocks)
}
func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//去取得数据块，发送
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)
	}
	if payload.Type == "tx" {
		txID := hex.EncodeToString([]byte(payload.ID))
		tx := mempool[txID]

		sendTx(payload.AddrFrom, &tx)
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	//存进交易池
	mempool[hex.EncodeToString(tx.ID)] = tx
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			//不是中心节点，和自身节点
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction
			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}
			//异常处理
			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, node := range knownNodes {
				//不是中心节点
				if node != nodeAddress {
					//广播出去
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			//内存还有继续循环
			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

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

	//添加为未知节点
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])

	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}
