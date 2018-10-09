package main

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
type getblocks struct {
	AddrFrom string
}
type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}
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

//?
func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}
