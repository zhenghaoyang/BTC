package pipieline

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"time"
)

var startTime time.Time
func Init() {
	startTime = time.Now()
}
//向通道读数据
//<- chan int 外部而言只读，对函数内部只写
func ArrarySource(a []int) <-chan int {
	//声明通道
	out := make(chan int)
	//开启goroutin，向通道传送数据
	go func() {
		for _, v := range a {
			//发送数据进chan
			out <- v
		}
		//关闭通道
		close(out)
	}()
	//返回通道
	return out
}

//读到内存排序后输出
func InMemSort(in <-chan int) <-chan int {
	//out := make(chan int)
	//优化
	out := make(chan int,1024)
	//读数据到内存
	go func() {
		arr := []int{}
		for v := range in {
			arr = append(arr, v)
		}
		fmt.Println("Read Data Done ",time.Now().Sub(startTime))
		sort.Ints(arr)
		fmt.Println("InMemSort Data Done ",time.Now().Sub(startTime))
		for _, v := range arr {
			out <- v
		}
		close(out)

	}()

	return out
}

//合并两个管道数据
func Merge(in1, in2 <-chan int) <-chan int {
	//out := make(chan int)
	//优化
	out := make(chan int,1024)
	go func() {
		v1, ok1 := <-in1
		v2, ok2 := <-in2
		//对管道两个管道数据排序
		for ok1 || ok2 { //不为false 表示有数据
			if !ok2 || (ok1 && v1 <= v2) { //in2没有数据，或者in1当前数据小于v2
				out <- v1
				//更新管道数据状态
				v1, ok1 = <-in1
			} else { //反之就是in2有数据,或者当前 v2<v1
				out <- v2
				v2, ok2 = <-in2
			}
		}
		close(out)
		fmt.Println("Merge Data Done ",time.Now().Sub(startTime))
	}()
	return out
}

//从文件里，向通道读数据 chunSize 分块大小
func ReaderSource(reader io.Reader, chunSize int) <-chan int {
	//out := make(chan int)
	//优化
	out := make(chan int,1024)
	go func() {
		buffer := make([]byte, 8)
		readBytes := 0
		for {
			//n, err :=bufio.NewReader(reader).Read(buffer)
			n, err := reader.Read(buffer)
			readBytes += n
			if n > 0 {
				v := int(binary.BigEndian.Uint64(buffer))
				out <- v
			}
			//-1 可以永远读下去，表示全部读， 不设-1最多只能读readBytes
			if err != nil || (chunSize != -1 && readBytes >= chunSize) {
				break
			}
		}
		close(out)
		//fmt.Println("Read Data Done ",time.Now().Sub(startTime))
	}()
	return out
}

func WriteSink(writer io.Writer, in <-chan int) {
	for v := range in {
		buffer := make([]byte, 8)
		binary.BigEndian.PutUint64(buffer, uint64(v))
		//bufio.NewWriter(writer).Write(buffer)
		writer.Write(buffer)
	}

}

func RandomSource(count int) <-chan int {
	out := make(chan int)
	go func() {
		for i := 0; i < count; i++ {
			out <- rand.Int()
		}
		close(out)
	}()

	return out
}

func MergeN(inputs ... <-chan int) <-chan int {
	//递归结束条件
	if len(inputs) == 1 {
		return inputs[0]
	}
	//靠近递归结束的条件
	m := len(inputs) / 2
	//递归处理两边数据
	return Merge(MergeN(inputs[:m]...), MergeN(inputs[m:]...))

}

//pratice
func ArrarySource1(a ...int) chan int {
	out := make(chan int)
	go func() {
		for _, v := range a {
			out <- v
		}
		close(out)
	}()
	return out
}

//函数参数只进不出 （只读）    只出不进（只写）
func MomorySort(in <-chan int) <-chan int { //只写
	out := make(chan int)
	go func() {
		s := []int{}
		for value := range in {
			//从in中读取数据
			//value = <- in  多读取一次 range 函数已经是读取
			s = append(s, value)
		}

		sort.Ints(s)

		for _, v := range s {
			//发送数据进out
			out <- v
		}
		close(out)
	}()

	return out
}
func Merge1(in1, in2 <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		v1, ok1 := <-in1
		v2, ok2 := <-in2
		for ok1 || ok2 {
			if !ok2 || (ok1 && v1 <= v2) {
				out <- v1
				v1, ok1 = <-in1
			} else {
				out <- v2
				v2, ok2 = <-in2
			}
		}
		close(out)
	}()
	return out
}
