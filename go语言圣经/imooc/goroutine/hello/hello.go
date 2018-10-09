package hello

import (
	"fmt"
)

func main() {
	ch := make(chan string)
	for i:= 0; i < 5 ;i++  {
		go printHello(i,ch)
	}
	for{
		msg := <- ch
		fmt.Println(msg)
	}
	//time.Sleep(time.Millisecond)
}
func printHello(i int,ch chan string){
		//ch = make(chan int)
		for j:= 10;j<=10;j++{
			ch <- fmt.Sprintf("hello from goroutines %d\n",i)
		}


}
