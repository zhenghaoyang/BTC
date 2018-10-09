package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
//区块中有多少交易
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

//getblocks 意为 “给我看一下你有什么区块”
type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

//inv 来向其他节点展示当前节点有什么块和交易
//Type 字段表明了这是块还是交易
type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
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

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

func sendBlock(addr string, b *Block) {
	fmt.Printf("test %v 节点请求区块", addr)
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
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

//发送清单
func sendInv(address, kind string, items [][]byte) {
	//清单包含，节点地址，请求类型
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	//包装为inv类型请求
	request := append(commandToBytes("inv"), payload...)
	//发送请求
	sendData(address, request)
}

func sendGetBlocks(address string) {
	//载荷
	payload := gobEncode(getblocks{nodeAddress})
	//包装为getblocks 类型的请求
	request := append(commandToBytes("getblocks"), payload...)
	//发送请求
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

//当接收到一个新块时，我们把它放到区块链里面。
//如果还有更多的区块需要下载，我们继续从上一个下载的块的那个节点继续请求。
//当最后把所有块都下载完后，对 UTXO 集进行重新索引。
func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block
	//跳过request type
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	//将数据解码进payload
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//取得payload中的Block
	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
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

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
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
	blocks := bc.GetBlockHashes()
	//发送清单
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

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
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
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
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

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

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

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
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
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		//处理连接
		go handleConnection(conn, bc)
	}
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

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}
