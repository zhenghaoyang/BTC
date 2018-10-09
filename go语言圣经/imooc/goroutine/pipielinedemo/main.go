package main

import (
	"BTC/go语言圣经/imooc/goroutine/pipieline"
	"bufio"
	"fmt"
	"os"
)

func main() {

	const filename = "large.in"
	const n = 64000000
	file, err := os.Create(filename)

	if err != nil {
		panic(err)
	}
	defer file.Close()
	out := pipieline.RandomSource(n)
	writer := bufio.NewWriter(file)
	pipieline.WriteSink(writer, out)

	//flush，不管多少，把缓存中的数据一次刷入到磁盘
	//write只有在满的时候才会刷入磁盘
	writer.Flush()

	file, err = os.Open(filename)
	if err != nil {
		panic(err)
	}
	in := pipieline.ReaderSource(bufio.NewReader(file),-1)

	count := 0
	for v := range in {
		fmt.Println(v)
		count ++
		if count >= 10 {
			break
		}
	}
}


func MergeDemo() {
	a := []int{2, 5, 9, 8, 7, 6}
	b := []int{3, 12, 34, 8, 1, 0}
	//out :=pipieline.ArrarySource(2,5,9,8,7,6)
	//out :=pipieline.InMemSort(pipieline.ArrarySource(b))
	out := pipieline.Merge(pipieline.InMemSort(pipieline.ArrarySource(a)), pipieline.InMemSort(pipieline.ArrarySource(b)))
	//out := pipieline.Merge1(pipieline.InMemSort(pipieline.ArrarySource(a)),pipieline.InMemSort(pipieline.ArrarySource(b)))
	//out := pipieline.InMemSort(pipieline.ArrarySource(a))
	//out := pipieline.MomorySort(pipieline.ArrarySource(a))
	for v := range out {
		fmt.Println(v)
	}

	//程序执行流程，
	//在主函数的range中等待其他goroutin的chan数据
	//主函数的range chan 等待 MomorySort开启的goroutin中的chan 数据
	// MomorySort开启的goroutin中的chan 等待 ArrarySource goroutin的chan的数据

	//first way
	//for  {
	//	//读取完毕，ok变为false
	//	if num , ok := <- out; ok{
	//		fmt.Println(num)
	//	} else{
	//		break
	//	}
	//}

}
