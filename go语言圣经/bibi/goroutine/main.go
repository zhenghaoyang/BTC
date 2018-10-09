package goroutine

import (
	"fmt"
	"runtime"
	"time"
)

//非抢占式多任务处理
//编译器/虚拟机/
//
func main()  {
	var a [10]int

	for i:=0;i<10 ;i++ {
		//data race
		//go func(){
		//	for {
		//		//闭包 //当 a=10 协程继续读取数据 //index out of range
		//		a[i]++
		//		runtime.Gosched()//让协程让出控制权
		//	}
		//}()
		go func(i int){
			for {
				//闭包 //当 a=10 协程继续读取数据 //index out of range
				a[i]++
				runtime.Gosched()
			}
		}(i)
		//go func(i int){
		//	for   {
		//		a[i]++
		//		//fmt.Println("hhhhh",i)//IO操作会交出控制权
				//没有下面这条语句,协程就不交出控制权,死循环
		//		//runtime.Gosched()//让协程让出控制权
		//	}
		//}(i)


	}
	//main 函数也是个协程,上面的协程停止,main没机会执行
	time.Sleep(time.Millisecond)
	fmt.Println(a)//main一边打印,go一边写,race
}