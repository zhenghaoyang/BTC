// 练习 1.3： 做实验测量潜在低效的版本+=和使用了strings.Join的版本的运行时间差异。
package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	p := fmt.Println
	str := ""
	begin := time.Now()

	for i := 0; i < 200000; i++ {
		str += "1"
	}
	end := time.Now()
	diff := end.Sub(begin)
	p(diff)
	var str2 []string
	begin2 := time.Now()

	for i := 0; i < 200000; i++ {
		strings.Join(str2, "1")
	}
	end2 := time.Now()
	diff = end2.Sub(begin2)
	p(diff)
}
