// 修改dup2，出现重复的行时打印文件名称
package main

// fmt.Printf("file neme =%s %d\t%s\n", os.Args[:1], n, line)
import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	counts := make(map[string]int)
	files := os.Args[:1]
	if len(files) == 0 {
		countLines(os.Stdin, counts)
	} else {
		for _, arg := range files {
			f, err := os.Open(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "dup2 %v \n", err)
				continue
			}
			countLines(f, counts)
			f.Close()
		}
	}
	for line, n := range counts {
		if n > 1 {
			// fmt.Printf("%d\t%s\n", n, line)
			fmt.Printf("file neme =%s %d\t%s\n", os.Args[:1], n, line)
		}
	}
}
func countLines(f *os.File, counts map[string]int) {
	input := bufio.NewScanner(f)
	for input.Scan() {
		counts[input.Text()]++
	}
}
