// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"os"
// )

// func main() {
// 	counts := make(map[string]int)
// 	files := os.Args[1:]
// 	if len(files) == 0 {
// 		countLines(os.Stdin, counts)
// 	} else {
// 		for _, arg := range files {
// 			// os.Open函数返回两个值。第一个值是被打开的文件(*os.File），
// 			// 其后被Scanner读取。
// 			f, err := os.Open(arg)
// 			if err != nil {
// 				fmt.Fprintf(os.Stderr, "dup2: %v\n", err)
// 				continue
// 			}
// 			//
// 			countLines(f, counts)
// 			f.Close()
// 		}
// 	}
// 	for line, n := range counts {
// 		if n > 1 {
// 			fmt.Printf("%d\t%s\n", n, line)
// 		}
// 	}
// }

// // countLines函数在其声明前被调用。
// // 函数和包级别的变量（package-level entities）可以任意顺序声明，并不影响其被调用。
// // map 引用传递
// // map作为为参数传递给某函数时，该函数接收这个引用的一份拷贝（copy，或译为副本），
// // 被调用函数对map底层数据结构的任何修改，调用者函数都可以通过持有的map引用看到。

// func countLines(f *os.File, counts map[string]int) {
// 	input := bufio.NewScanner(f)
// 	//input.Scan读取文件
// 	for input.Scan() {
// 		counts[input.Text()]++
// 	}
// }
