package main

import (
	"fmt"
	"os"
)

func main() {
	//修改echo程序，使其能够打印os.Args[0]，即被执行命令本身的名字。
	// 修改echo程序，使其打印每个参数的索引和值，每个一行。
	for k, arg := range os.Args[0:] {
		fmt.Print("k=", k)
		fmt.Print(" arg=", arg)
		fmt.Println()
	}

}
