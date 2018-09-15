package main

type Blockchain struct {
	//区块数组切片
	blocks []*Block
}

//AddBlock 添加区块
func (bc *Blockchain) AddBlock(data string) {
	//找到前一个区块
	preBlock := bc.blocks[len(bc.blocks)-1]
	//创建新区块
	newBlock := NewBlock(data, preBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

func NewBlockchain() *Blockchain {
	//区块链
	return &Blockchain{
		//创始区块
		[]*Block{NewGenesisBlock()},
	}
}
