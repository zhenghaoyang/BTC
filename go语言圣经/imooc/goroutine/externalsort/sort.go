package main

import (
	"BTC/go语言圣经/imooc/goroutine/pipieline"
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main() {
	//p := CreatePipeline("large.in", 512000000, 4)
	//p := CreateNetWorkPipeline("small.in", 512, 4)
	p := CreateNetWorkPipeline("large.in", 512000000, 4)
	//time.Sleep(time.Hour)
	writeToFile(p, "large.out")
	printFile("large.out")
}
func printFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	p := pipieline.ReaderSource(file, -1)
	count := 0
	for v := range p {
		fmt.Println(v)
		count++
		if count >= 10{
			break
		}
	}
}
func writeToFile(p <-chan int, filename string) {

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	pipieline.WriteSink(writer, p)

}

func CreatePipeline(
	filename string,
	fileSize, chunCount int, ) <-chan int {
	//初始化时间
	pipieline.Init()
	chunSize := fileSize / chunCount

	sourceResults := []<-chan int{}
	//把一个文件分成多块读进不同的chan
	for i := 0; i < chunCount; i++ {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		file.Seek(int64(i*chunSize), 0)
		source := pipieline.ReaderSource(bufio.NewReader(file), chunSize)
		//排序,添加到多个管道中
		sourceResults = append(sourceResults, pipieline.InMemSort(source))
	}
	return pipieline.MergeN(sourceResults...)
}


func CreateNetWorkPipeline(
	filename string,
	fileSize, chunCount int, ) <-chan int {
	//初始化时间
	pipieline.Init()
	chunSize := fileSize / chunCount

	//sourceResults := []<-chan int{}
	sortAddr := []string{}
	//把一个文件分成多块读进不同的chan
	for i := 0; i < chunCount; i++ {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		file.Seek(int64(i*chunSize), 0)
		source := pipieline.ReaderSource(bufio.NewReader(file), chunSize)

		addr := ":"+strconv.Itoa(7000+i)
		//server
		pipieline.NetworkSink(addr,pipieline.InMemSort(source))
		sortAddr = append(sortAddr,addr)
	}

	//测试server
	//return nil

	sourceResults := []<-chan int{}
	for _,addr := range sortAddr{
		//client 去连接上上面的server,读取数据
		sourceResults = append(sourceResults,pipieline.NetWorkSource(addr))
	}
	return pipieline.MergeN(sourceResults...)
}